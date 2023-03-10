package instruments

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/pslive/internal/app/pslive/auth"
	"github.com/sargassum-world/pslive/internal/app/pslive/handling"
	"github.com/sargassum-world/pslive/internal/clients/chat"
	"github.com/sargassum-world/pslive/internal/clients/instruments"
	"github.com/sargassum-world/pslive/internal/clients/ory"
	"github.com/sargassum-world/pslive/internal/clients/planktoscope"
	"github.com/sargassum-world/pslive/internal/clients/presence"
)

func parseID(raw string, typeName string) (int64, error) {
	const intBase = 10
	const intWidth = 64
	id, err := strconv.ParseInt(raw, intBase, intWidth)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid %s id", typeName))
	}
	return id, err
}

// Instrument

type InstrumentViewData struct {
	Instrument       instruments.Instrument
	ControllerIDs    []int64
	Controllers      map[int64]planktoscope.Planktoscope
	AdminIdentifier  string
	KnownViewers     []presence.User
	AnonymousViewers []string
	ChatMessages     []handling.ChatMessageViewData
}

func getInstrumentViewData(
	ctx context.Context, instrumentID int64,
	oc *ory.Client, is *instruments.Store, pco *planktoscope.Orchestrator,
	ps *presence.Store, cs *chat.Store,
) (vd InstrumentViewData, err error) {
	if vd.Instrument, err = is.GetInstrument(ctx, instrumentID); err != nil {
		// TODO: is this the best way to handle errors from is.GetInstrumentByID?
		return InstrumentViewData{}, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("instrument %d not found", instrumentID),
		)
	}

	vd.ControllerIDs = make([]int64, 0, len(vd.Instrument.Controllers))
	vd.Controllers = make(map[int64]planktoscope.Planktoscope)
	for _, controller := range vd.Instrument.Controllers {
		pc, ok := pco.Get(controller.ID)
		if !ok {
			return InstrumentViewData{}, errors.Errorf(
				"planktoscope client for instrument %d not found", instrumentID,
			)
		}
		if pc.HasConnection() {
			// TODO: display some indication to the user when a controller is unreachable, and push
			// updates over Turbo Streams when a controller's reachability changes
			vd.ControllerIDs = append(vd.ControllerIDs, controller.ID)
			vd.Controllers[controller.ID] = pc.GetState()
		}
	}

	if vd.AdminIdentifier, err = oc.GetIdentifier(ctx, vd.Instrument.AdminID); err != nil {
		return InstrumentViewData{}, errors.Wrapf(
			err, "couldn't look up admin identifier for instrument %d", instrumentID,
		)
	}

	// Chat
	vd.KnownViewers, vd.AnonymousViewers = ps.List(fmt.Sprintf("/instruments/%d/users", instrumentID))
	messages, err := cs.GetMessagesByTopic(
		ctx, fmt.Sprintf("/instruments/%d/chat/messages", instrumentID), chat.DefaultMessagesLimit,
	)
	if err != nil {
		return InstrumentViewData{}, errors.Wrapf(
			err, "couldn't get chat messages for instrument %d", instrumentID,
		)
	}
	vd.ChatMessages, err = handling.AdaptChatMessages(ctx, messages, oc)
	if err != nil {
		return InstrumentViewData{}, errors.Wrapf(
			err, "couldn't adapt chat messages for instrument %d into view data", instrumentID,
		)
	}

	return vd, nil
}

type InstrumentViewAuthz struct {
	SendChat    bool
	Controllers map[int64]interface{}
}

func getInstrumentViewAuthz(
	ctx context.Context, instrumentID int64, controllerIDs []int64, a auth.Auth,
	azc *auth.AuthzChecker,
) (authz InstrumentViewAuthz, err error) {
	eg, egctx := errgroup.WithContext(ctx)
	controllerAuthorizations := make([]interface{}, len(controllerIDs))
	for i, controllerID := range controllerIDs {
		eg.Go(func(i int, cid int64) func() error {
			return func() (err error) {
				if controllerAuthorizations[i], err = getPlanktoscopeControllerViewAuthz(
					egctx, instrumentID, cid, a, azc,
				); err != nil {
					return errors.Wrapf(
						err, "couldn't check authz for controller %d for instrument %d", cid, instrumentID,
					)
				}
				return nil
			}
		}(i, controllerID))
	}
	eg.Go(func() (err error) {
		path := fmt.Sprintf("/instruments/%d/chat/messages", instrumentID)
		if authz.SendChat, err = azc.Allow(egctx, a, path, http.MethodPost, nil); err != nil {
			return errors.Wrapf(
				err, "couldn't check authz for sending to chat for instrument %d", instrumentID,
			)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return InstrumentViewAuthz{}, err
	}
	authz.Controllers = make(map[int64]interface{})
	for i, controllerID := range controllerIDs {
		authz.Controllers[controllerID] = controllerAuthorizations[i]
	}
	return authz, nil
}

func (h *Handlers) HandleInstrumentGet() auth.HTTPHandlerFunc {
	t := "instruments/instrument.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		instrumentID, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}

		// Run queries
		ctx := c.Request().Context()
		instrumentViewData, err := getInstrumentViewData(
			ctx, instrumentID, h.oc, h.is, h.pco, h.ps, h.cs,
		)
		if err != nil {
			return err
		}
		if a.Authorizations, err = getInstrumentViewAuthz(
			ctx, instrumentID, instrumentViewData.ControllerIDs, a, h.azc,
		); err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, instrumentViewData, a)
	}
}

func (h *Handlers) HandleInstrumentPost() auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		instrumentID, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		state := c.FormValue("state")

		// Run queries
		ctx := c.Request().Context()
		switch state {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid instrument state %s", state,
			))
		case "deleted":
			// FIXME: there needs to be an authorization check to ensure that the user attempting to
			// delete the instrument is an administrator of the instrument!

			if err = h.is.DeleteInstrument(ctx, instrumentID); err != nil {
				return err
			}
			// TODO: cancel any relevant turbo streams topics

			// Redirect user
			return c.Redirect(http.StatusSeeOther, "/instruments")
		}
	}
}

func (h *Handlers) HandleInstrumentNamePost() auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		instrumentID, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		name := c.FormValue("name")

		// Run queries
		// FIXME: there needs to be an authorization check to ensure that the user attempting to
		// delete the instrument is an administrator of the instrument!
		if err := h.is.UpdateInstrumentName(c.Request().Context(), instrumentID, name); err != nil {
			return err
		}

		// TODO: return turbo stream, broadcast updates

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", instrumentID))
	}
}

func (h *Handlers) HandleInstrumentDescriptionPost() auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		instrumentID, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		description := c.FormValue("description")

		// Run queries
		// FIXME: there needs to be an authorization check to ensure that the user attempting to
		// delete the instrument is an administrator of the instrument!
		if err := h.is.UpdateInstrumentDescription(
			c.Request().Context(), instrumentID, description,
		); err != nil {
			return err
		}

		// TODO: return turbo stream, broadcast updates

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", instrumentID))
	}
}

// Components

func handleInstrumentComponentsPost(
	storeAdder func(ctx context.Context, instrumentID int64, url, protocol string) error,
) auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		instrumentID, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		url := c.FormValue("url")
		protocol := c.FormValue("protocol")

		// Run queries
		// FIXME: there needs to be an authorization check to ensure that the user attempting to
		// delete the instrument is an administrator of the instrument!
		if err := storeAdder(c.Request().Context(), instrumentID, url, protocol); err != nil {
			return err
		}

		// TODO: return turbo stream, broadcast updates

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", instrumentID))
	}
}

func handleInstrumentComponentPost(
	typeName string,
	componentUpdater func(ctx context.Context, componentID int64, url, protocol string) error,
	componentDeleter func(ctx context.Context, componentID int64) error,
) auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		instrumentID, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		componentID, err := parseID(c.Param(typeName+"ID"), typeName)
		if err != nil {
			return err
		}
		state := c.FormValue("state")

		// Run queries
		ctx := c.Request().Context()
		switch state {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid %s state %s", typeName, state,
			))
		case "updated":
			protocol := c.FormValue("protocol")
			url := c.FormValue("url")
			// FIXME: needs authorization check!
			if err = componentUpdater(ctx, componentID, url, protocol); err != nil {
				return err
			}
			// TODO: deal with turbo streams
		case "deleted":
			// FIXME: needs authorization check!
			if err = componentDeleter(ctx, componentID); err != nil {
				return err
			}
			// TODO: deal with turbo streams
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", instrumentID))
	}
}

// Controllers

func (h *Handlers) HandleInstrumentControllersPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentsPost(
		func(ctx context.Context, instrumentID int64, url, protocol string) error {
			controllerID, err := h.is.AddController(ctx, instruments.Controller{
				InstrumentID: instrumentID,
				URL:          url,
				Protocol:     protocol,
			})
			if err != nil {
				return err
			}
			if err := h.pco.Add(controllerID, url); err != nil {
				return err
			}
			return nil
		},
	)
}

func (h *Handlers) HandleInstrumentControllerPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentPost(
		"controller",
		func(ctx context.Context, componentID int64, url, protocol string) error {
			if err := h.is.UpdateController(ctx, instruments.Controller{
				ID:       componentID,
				URL:      url,
				Protocol: protocol,
			}); err != nil {
				return err
			}
			if err := h.pco.Update(ctx, componentID, url); err != nil {
				return err
			}
			return nil
		},
		func(ctx context.Context, componentID int64) error {
			if err := h.is.DeleteController(ctx, componentID); err != nil {
				return err
			}
			if err := h.pco.Remove(ctx, componentID); err != nil {
				return err
			}
			return nil
		},
	)
}
