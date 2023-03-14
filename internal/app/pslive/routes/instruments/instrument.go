package instruments

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

const flagChecked = "true"

func parseID[ID ~int64](raw string, typeName string) (ID, error) {
	const intBase = 10
	const intWidth = 64
	id, err := strconv.ParseInt(raw, intBase, intWidth)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
			"invalid %s id %s", typeName, raw,
		))
	}
	return ID(id), err
}

// Components (common across cameras & controllers)

func handleInstrumentComponentsPost(
	storeAdder func(
		ctx context.Context, iid instruments.InstrumentID,
		enabled bool, name, description string, params url.Values,
	) error,
) auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		iid, err := parseID[instruments.InstrumentID](c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		enabled := strings.ToLower(c.FormValue("enabled")) == flagChecked
		name := c.FormValue("name")
		description := c.FormValue("description")
		params, err := c.FormParams()
		if err != nil {
			return errors.Wrap(err, "couldn't parse form params")
		}

		// Run queries
		if err := storeAdder(
			c.Request().Context(), iid, enabled, name, description, params,
		); err != nil {
			return err
		}

		// TODO: return turbo stream, broadcast updates

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", iid))
	}
}

func handleInstrumentComponentPost[ComponentID ~int64](
	typeName string,
	componentUpdater func(
		ctx context.Context, componentID ComponentID, instrumentID instruments.InstrumentID,
		enabled bool, name, description string, params url.Values,
	) error,
	componentDeleter func(ctx context.Context, componentID ComponentID) error,
) auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		iid, err := parseID[instruments.InstrumentID](c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		componentID, err := parseID[ComponentID](c.Param(typeName+"ID"), typeName)
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
			enabled := strings.ToLower(c.FormValue("enabled")) == flagChecked
			name := c.FormValue("name")
			description := c.FormValue("description")
			params, perr := c.FormParams()
			if perr != nil {
				return errors.Wrap(err, "couldn't parse form params")
			}
			if err = componentUpdater(
				ctx, componentID, iid, enabled, name, description, params,
			); err != nil {
				return err
			}
			// TODO: deal with turbo streams
		case "deleted":
			if err = componentDeleter(ctx, componentID); err != nil {
				return err
			}
			// TODO: deal with turbo streams
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", iid))
	}
}

// Instrument

type InstrumentViewData struct {
	Instrument       instruments.Instrument
	ControllerIDs    []instruments.ControllerID
	Controllers      map[instruments.ControllerID]planktoscope.Planktoscope
	AdminIdentifier  ory.IdentityIdentifier
	KnownViewers     []presence.User
	AnonymousViewers []presence.SessionID
	ChatMessages     []handling.ChatMessageViewData
}

func getInstrumentViewData(
	ctx context.Context, iid instruments.InstrumentID,
	oc *ory.Client, is *instruments.Store, pco *planktoscope.Orchestrator,
	ps *presence.Store, cs *chat.Store,
) (vd InstrumentViewData, err error) {
	if vd.Instrument, err = is.GetInstrument(ctx, iid); err != nil {
		// TODO: is this the best way to handle errors from is.GetInstrumentByID?
		return InstrumentViewData{}, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("instrument %d not found", iid),
		)
	}

	vd.ControllerIDs = make([]instruments.ControllerID, 0, len(vd.Instrument.Controllers))
	vd.Controllers = make(map[instruments.ControllerID]planktoscope.Planktoscope)
	for _, controller := range vd.Instrument.Controllers {
		if !controller.Enabled {
			continue
		}
		pc, ok := pco.Get(planktoscope.ClientID(controller.ID))
		if !ok {
			return InstrumentViewData{}, errors.Errorf(
				"planktoscope client for instrument %d not found", iid,
			)
		}
		if pc.HasConnection() {
			// TODO: display some indication to the user when a controller is unreachable, and push
			// updates over Turbo Streams when a controller's reachability changes
			vd.ControllerIDs = append(vd.ControllerIDs, controller.ID)
			vd.Controllers[controller.ID] = pc.GetState()
		}
	}

	if vd.AdminIdentifier, err = oc.GetIdentifier(
		ctx, ory.IdentityID(vd.Instrument.AdminID),
	); err != nil {
		return InstrumentViewData{}, errors.Wrapf(
			err, "couldn't look up admin identifier for instrument %d", iid,
		)
	}

	// Chat
	vd.KnownViewers, vd.AnonymousViewers = ps.List(presence.Topic(
		fmt.Sprintf("/instruments/%d/users", iid)))
	messages, err := cs.GetMessagesByTopic(
		ctx, chat.Topic(fmt.Sprintf("/instruments/%d/chat/messages", iid)),
		chat.DefaultMessagesLimit,
	)
	if err != nil {
		return InstrumentViewData{}, errors.Wrapf(
			err, "couldn't get chat messages for instrument %d", iid,
		)
	}
	vd.ChatMessages, err = handling.AdaptChatMessages(ctx, messages, oc)
	if err != nil {
		return InstrumentViewData{}, errors.Wrapf(
			err, "couldn't adapt chat messages for instrument %d into view data", iid,
		)
	}

	return vd, nil
}

type InstrumentViewAuthz struct {
	SendChat    bool
	Controllers map[instruments.ControllerID]interface{}
}

func getInstrumentViewAuthz(
	ctx context.Context, iid instruments.InstrumentID, controllerIDs []instruments.ControllerID,
	a auth.Auth, azc *auth.AuthzChecker,
) (authz InstrumentViewAuthz, err error) {
	eg, egctx := errgroup.WithContext(ctx)
	controllerAuthorizations := make([]interface{}, len(controllerIDs))
	for i, controllerID := range controllerIDs {
		eg.Go(func(i int, cid instruments.ControllerID) func() error {
			return func() (err error) {
				if controllerAuthorizations[i], err = getPlanktoscopeControllerViewAuthz(
					egctx, iid, cid, a, azc,
				); err != nil {
					return errors.Wrapf(
						err, "couldn't check authz for controller %d for instrument %d", cid, iid,
					)
				}
				return nil
			}
		}(i, controllerID))
	}
	eg.Go(func() (err error) {
		path := fmt.Sprintf("/instruments/%d/chat/messages", iid)
		if authz.SendChat, err = azc.Allow(egctx, a, path, http.MethodPost, nil); err != nil {
			return errors.Wrapf(
				err, "couldn't check authz for sending to chat for instrument %d", iid,
			)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return InstrumentViewAuthz{}, err
	}
	authz.Controllers = make(map[instruments.ControllerID]interface{})
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
		iid, err := parseID[instruments.InstrumentID](c.Param("id"), "instrument")
		if err != nil {
			return err
		}

		// Run queries
		ctx := c.Request().Context()
		instrumentViewData, err := getInstrumentViewData(ctx, iid, h.oc, h.is, h.pco, h.ps, h.cs)
		if err != nil {
			return err
		}
		if a.Authorizations, err = getInstrumentViewAuthz(
			ctx, iid, instrumentViewData.ControllerIDs, a, h.azc,
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
		iid, err := parseID[instruments.InstrumentID](c.Param("id"), "instrument")
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
			if err = h.is.DeleteInstrument(ctx, iid); err != nil {
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
		iid, err := parseID[instruments.InstrumentID](c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		name := c.FormValue("name")

		// Run queries
		if err := h.is.UpdateInstrumentName(c.Request().Context(), iid, name); err != nil {
			return err
		}

		// TODO: return turbo stream, broadcast updates

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", iid))
	}
}

func (h *Handlers) HandleInstrumentDescriptionPost() auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		iid, err := parseID[instruments.InstrumentID](c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		description := c.FormValue("description")

		// Run queries
		if err := h.is.UpdateInstrumentDescription(
			c.Request().Context(), iid, description,
		); err != nil {
			return err
		}

		// TODO: return turbo stream, broadcast updates

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", iid))
	}
}
