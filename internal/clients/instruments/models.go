package instruments

import (
	"zombiezen.com/go/sqlite"
)

// Camera

type Camera struct {
	ID           int64
	InstrumentID int64
	URL          string
	Protocol     string
}

func (c Camera) newInsertion() map[string]interface{} {
	return map[string]interface{}{
		"$url":           c.URL,
		"$protocol":      c.Protocol,
		"$instrument_id": c.InstrumentID,
	}
}

func (c Camera) newUpdate() map[string]interface{} {
	return map[string]interface{}{
		"$id":       c.ID,
		"$url":      c.URL,
		"$protocol": c.Protocol,
	}
}

func (c Camera) newDelete() map[string]interface{} {
	return map[string]interface{}{
		"$id": c.ID,
	}
}

// Controller

type Controller struct {
	ID           int64
	InstrumentID int64
	URL          string
	Protocol     string
}

func (c Controller) newInsertion() map[string]interface{} {
	return map[string]interface{}{
		"$url":           c.URL,
		"$protocol":      c.Protocol,
		"$instrument_id": c.InstrumentID,
	}
}

func (c Controller) newUpdate() map[string]interface{} {
	return map[string]interface{}{
		"$id":       c.ID,
		"$url":      c.URL,
		"$protocol": c.Protocol,
	}
}

func (c Controller) newDelete() map[string]interface{} {
	return map[string]interface{}{
		"$id": c.ID,
	}
}

func (c Controller) newProtocolSelection() map[string]interface{} {
	return map[string]interface{}{
		"$protocol": c.Protocol,
	}
}

// Controllers

type controllersSelector struct {
	ids         []int64
	controllers map[int64]Controller
}

func newControllersSelector() *controllersSelector {
	return &controllersSelector{
		ids:         make([]int64, 0),
		controllers: make(map[int64]Controller),
	}
}

func (sel *controllersSelector) Step(s *sqlite.Stmt) error {
	id := s.GetInt64("id")
	if _, ok := sel.controllers[id]; !ok {
		sel.controllers[id] = Controller{
			ID:           s.GetInt64("id"),
			InstrumentID: s.GetInt64("instrument_id"),
			URL:          s.GetText("url"),
			Protocol:     s.GetText("protocol"),
		}
		if id != 0 {
			sel.ids = append(sel.ids, id)
		}
	}
	return nil
}

func (sel *controllersSelector) Controllers() []Controller {
	controllers := make([]Controller, len(sel.ids))
	for i, id := range sel.ids {
		controllers[i] = sel.controllers[id]
	}
	return controllers
}

// Instrument

type Instrument struct {
	ID          int64
	Name        string
	Description string
	AdminID     string
	Cameras     map[int64]Camera
	Controllers map[int64]Controller
}

func (i Instrument) newInsertion() map[string]interface{} {
	return map[string]interface{}{
		"$name":        i.Name,
		"$description": i.Description,
		"$admin_id":    i.AdminID,
	}
}

func (i Instrument) newNameUpdate() map[string]interface{} {
	return map[string]interface{}{
		"$id":   i.ID,
		"$name": i.Name,
	}
}

func (i Instrument) newDescriptionUpdate() map[string]interface{} {
	return map[string]interface{}{
		"$id":          i.ID,
		"$description": i.Description,
	}
}

func (i Instrument) newDelete() map[string]interface{} {
	return map[string]interface{}{
		"$id": i.ID,
	}
}

func (i Instrument) newAdminIDSelection() map[string]interface{} {
	return map[string]interface{}{
		"$admin_id": i.AdminID,
	}
}

// Instruments

type instrumentsSelector struct {
	ids         []int64
	instruments map[int64]Instrument
}

func newInstrumentsSelector() *instrumentsSelector {
	return &instrumentsSelector{
		ids:         make([]int64, 0),
		instruments: make(map[int64]Instrument),
	}
}

func (sel *instrumentsSelector) Step(s *sqlite.Stmt) error {
	id := s.GetInt64("id")
	if _, ok := sel.instruments[id]; !ok {
		sel.instruments[id] = Instrument{
			ID:          s.GetInt64("id"),
			Name:        s.GetText("name"),
			Description: s.GetText("description"),
			AdminID:     s.GetText("admin_id"),
			Cameras:     make(map[int64]Camera),
			Controllers: make(map[int64]Controller),
		}
		if id != 0 {
			sel.ids = append(sel.ids, id)
		}
	}
	instrument := sel.instruments[id]

	cameraID := s.GetInt64("camera_id")
	camera := Camera{
		ID:           cameraID,
		InstrumentID: s.GetInt64("id"),
		URL:          s.GetText("camera_url"),
		Protocol:     s.GetText("camera_protocol"),
	}
	if camera != (Camera{
		ID:           cameraID,
		InstrumentID: id,
	}) {
		instrument.Cameras[cameraID] = camera
	}

	controllerID := s.GetInt64("controller_id")
	controller := Controller{
		ID:           controllerID,
		InstrumentID: s.GetInt64("id"),
		URL:          s.GetText("controller_url"),
		Protocol:     s.GetText("controller_protocol"),
	}
	if controller != (Controller{
		ID:           controllerID,
		InstrumentID: id,
	}) {
		instrument.Controllers[controllerID] = controller
	}

	sel.instruments[id] = instrument
	return nil
}

func (sel *instrumentsSelector) Instruments() []Instrument {
	instruments := make([]Instrument, len(sel.ids))
	for i, id := range sel.ids {
		instruments[i] = sel.instruments[id]
	}
	return instruments
}
