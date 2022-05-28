package instruments

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

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
	ctx context.Context, id int64,
	oc *ory.Client, is *instruments.Store, pco *planktoscope.Orchestrator,
	ps *presence.Store, cs *chat.Store,
) (vd InstrumentViewData, err error) {
	if vd.Instrument, err = is.GetInstrument(ctx, id); err != nil {
		// TODO: is this the best way to handle errors from is.GetInstrumentByID?
		return InstrumentViewData{}, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("instrument %d not found", id),
		)
	}

	vd.ControllerIDs = make([]int64, 0, len(vd.Instrument.Controllers))
	vd.Controllers = make(map[int64]planktoscope.Planktoscope)
	for _, controller := range vd.Instrument.Controllers {
		pc, ok := pco.Get(controller.ID)
		if !ok {
			return InstrumentViewData{}, errors.Errorf(
				"planktoscope client for instrument %d not found", id,
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
			err, "couldn't look up admin identifier for instrument %d", id,
		)
	}

	// Chat
	vd.KnownViewers, vd.AnonymousViewers = ps.List(fmt.Sprintf("/instruments/%d/users", id))
	messages, err := cs.GetMessagesByTopic(
		ctx, fmt.Sprintf("/instruments/%d/chat/messages", id), chat.DefaultMessagesLimit,
	)
	if err != nil {
		return InstrumentViewData{}, errors.Wrapf(
			err, "couldn't get chat messages for instrument %d", id,
		)
	}
	vd.ChatMessages, err = handling.AdaptChatMessages(ctx, messages, oc)
	if err != nil {
		return InstrumentViewData{}, errors.Wrapf(
			err, "couldn't adapt chat messages for instrument %d into view data", id,
		)
	}

	return vd, nil
}

func (h *Handlers) HandleInstrumentGet() auth.HTTPHandlerFunc {
	t := "instruments/instrument.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}

		// Run queries
		instrumentViewData, err := getInstrumentViewData(
			c.Request().Context(), id, h.oc, h.is, h.pco, h.ps, h.cs,
		)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, instrumentViewData, a)
	}
}

func (h *Handlers) HandleInstrumentPost() auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id, err := parseID(c.Param("id"), "instrument")
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

			if err = h.is.DeleteInstrument(ctx, id); err != nil {
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
		id, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		name := c.FormValue("name")

		// Run queries
		// FIXME: there needs to be an authorization check to ensure that the user attempting to
		// delete the instrument is an administrator of the instrument!
		if err := h.is.UpdateInstrumentName(c.Request().Context(), id, name); err != nil {
			return err
		}

		// TODO: return turbo stream, broadcast updates

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", id))
	}
}

func (h *Handlers) HandleInstrumentDescriptionPost() auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		description := c.FormValue("description")

		// Run queries
		// FIXME: there needs to be an authorization check to ensure that the user attempting to
		// delete the instrument is an administrator of the instrument!
		if err := h.is.UpdateInstrumentDescription(c.Request().Context(), id, description); err != nil {
			return err
		}

		// TODO: return turbo stream, broadcast updates

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", id))
	}
}

// Components

func handleInstrumentComponentsPost(
	storeAdder func(ctx context.Context, id int64, url, protocol string) error,
) auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id, err := parseID(c.Param("id"), "instrument")
		if err != nil {
			return err
		}
		url := c.FormValue("url")
		protocol := c.FormValue("protocol")

		// Run queries
		// FIXME: there needs to be an authorization check to ensure that the user attempting to
		// delete the instrument is an administrator of the instrument!
		if err := storeAdder(c.Request().Context(), id, url, protocol); err != nil {
			return err
		}

		// TODO: return turbo stream, broadcast updates

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", id))
	}
}

func handleInstrumentComponentPost(
	typeName string,
	componentUpdater func(ctx context.Context, componentID int64, url, protocol string) error,
	componentDeleter func(ctx context.Context, componentID int64) error,
) auth.HTTPHandlerFunc {
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id, err := parseID(c.Param("id"), "instrument")
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
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/instruments/%d", id))
	}
}

// Cameras

func (h *Handlers) HandleInstrumentCamerasPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentsPost(
		func(ctx context.Context, id int64, url, protocol string) error {
			_, err := h.is.AddCamera(ctx, instruments.Camera{
				InstrumentID: id,
				URL:          url,
				Protocol:     protocol,
			})
			return err
		},
	)
}

func (h *Handlers) HandleInstrumentCameraPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentPost(
		"camera",
		func(ctx context.Context, componentID int64, url, protocol string) error {
			return h.is.UpdateCamera(ctx, instruments.Camera{
				ID:       componentID,
				URL:      url,
				Protocol: protocol,
			})
		},
		h.is.DeleteCamera,
	)
}

// Controllers

func (h *Handlers) HandleInstrumentControllersPost() auth.HTTPHandlerFunc {
	return handleInstrumentComponentsPost(
		func(ctx context.Context, id int64, url, protocol string) error {
			controllerID, err := h.is.AddController(ctx, instruments.Controller{
				InstrumentID: id,
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
