package instruments

import (
	"zombiezen.com/go/sqlite"
)

type (
	CameraID     int64
	ControllerID int64
	InstrumentID int64
	AdminID      string
)

// Camera

type Camera struct {
	ID           CameraID
	InstrumentID InstrumentID
	URL          string
	Protocol     string
	Enabled      bool
}

func (c Camera) newInsertion() map[string]interface{} {
	return map[string]interface{}{
		"$instrument_id": c.InstrumentID,
		"$url":           c.URL,
		"$protocol":      c.Protocol,
		"$enabled":       c.Enabled,
	}
}

func (c Camera) newUpdate() map[string]interface{} {
	return map[string]interface{}{
		"$id":       c.ID,
		"$url":      c.URL,
		"$protocol": c.Protocol,
		"$enabled":  c.Enabled,
	}
}

func (c Camera) newDelete() map[string]interface{} {
	return map[string]interface{}{
		"$id": c.ID,
	}
}

func newCameraSelection(id CameraID) map[string]interface{} {
	return map[string]interface{}{
		"$id": id,
	}
}

// Cameras

type camerasSelector struct {
	ids     []CameraID
	cameras map[CameraID]Camera
}

func newCamerasSelector() *camerasSelector {
	return &camerasSelector{
		ids:     make([]CameraID, 0),
		cameras: make(map[CameraID]Camera),
	}
}

func (sel *camerasSelector) Step(s *sqlite.Stmt) error {
	id := CameraID(s.GetInt64("id"))
	if _, ok := sel.cameras[id]; !ok {
		sel.cameras[id] = Camera{
			ID:           id,
			InstrumentID: InstrumentID(s.GetInt64("instrument_id")),
			URL:          s.GetText("url"),
			Protocol:     s.GetText("protocol"),
			Enabled:      s.GetBool("enabled"),
		}
		if id != 0 {
			sel.ids = append(sel.ids, id)
		}
	}
	return nil
}

func (sel *camerasSelector) Cameras() []Camera {
	cameras := make([]Camera, len(sel.ids))
	for i, id := range sel.ids {
		cameras[i] = sel.cameras[id]
	}
	return cameras
}

// Controller

type Controller struct {
	ID           ControllerID
	InstrumentID InstrumentID
	URL          string
	Protocol     string
	Enabled      bool
}

func (c Controller) newInsertion() map[string]interface{} {
	return map[string]interface{}{
		"$instrument_id": c.InstrumentID,
		"$url":           c.URL,
		"$protocol":      c.Protocol,
		"$enabled":       c.Enabled,
	}
}

func (c Controller) newUpdate() map[string]interface{} {
	return map[string]interface{}{
		"$id":       c.ID,
		"$url":      c.URL,
		"$protocol": c.Protocol,
		"$enabled":  c.Enabled,
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
	ids         []ControllerID
	controllers map[ControllerID]Controller
}

func newControllersSelector() *controllersSelector {
	return &controllersSelector{
		ids:         make([]ControllerID, 0),
		controllers: make(map[ControllerID]Controller),
	}
}

func (sel *controllersSelector) Step(s *sqlite.Stmt) error {
	id := ControllerID(s.GetInt64("id"))
	if _, ok := sel.controllers[id]; !ok {
		sel.controllers[id] = Controller{
			ID:           id,
			InstrumentID: InstrumentID(s.GetInt64("instrument_id")),
			URL:          s.GetText("url"),
			Protocol:     s.GetText("protocol"),
			Enabled:      s.GetBool("enabled"),
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
	ID          InstrumentID
	Name        string
	Description string
	AdminID     AdminID
	Cameras     map[CameraID]Camera
	Controllers map[ControllerID]Controller
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

func newInstrumentSelection(id InstrumentID) map[string]interface{} {
	return map[string]interface{}{
		"$id": id,
	}
}

// Instruments

func newInstrumentsSelection() map[string]interface{} {
	return map[string]interface{}{}
}

type instrumentsSelector struct {
	ids         []InstrumentID
	instruments map[InstrumentID]Instrument
}

func newInstrumentsSelector() *instrumentsSelector {
	return &instrumentsSelector{
		ids:         make([]InstrumentID, 0),
		instruments: make(map[InstrumentID]Instrument),
	}
}

func (sel *instrumentsSelector) Step(s *sqlite.Stmt) error {
	instrumentID := InstrumentID(s.GetInt64("id"))
	if _, ok := sel.instruments[instrumentID]; !ok {
		sel.instruments[instrumentID] = Instrument{
			ID:          instrumentID,
			Name:        s.GetText("name"),
			Description: s.GetText("description"),
			AdminID:     AdminID(s.GetText("admin_id")),
			Cameras:     make(map[CameraID]Camera),
			Controllers: make(map[ControllerID]Controller),
		}
		if instrumentID != 0 {
			sel.ids = append(sel.ids, instrumentID)
		}
	}
	instrument := sel.instruments[instrumentID]

	cameraID := CameraID(s.GetInt64("camera_id"))
	camera := Camera{
		ID:           cameraID,
		InstrumentID: instrumentID,
		URL:          s.GetText("camera_url"),
		Protocol:     s.GetText("camera_protocol"),
		Enabled:      s.GetBool("camera_enabled"),
	}
	if camera != (Camera{
		ID:           cameraID,
		InstrumentID: instrumentID,
	}) {
		instrument.Cameras[cameraID] = camera
	}

	controllerID := ControllerID(s.GetInt64("controller_id"))
	controller := Controller{
		ID:           controllerID,
		InstrumentID: instrumentID,
		URL:          s.GetText("controller_url"),
		Protocol:     s.GetText("controller_protocol"),
		Enabled:      s.GetBool("controller_enabled"),
	}
	if controller != (Controller{
		ID:           controllerID,
		InstrumentID: instrumentID,
	}) {
		instrument.Controllers[controllerID] = controller
	}

	sel.instruments[instrumentID] = instrument
	return nil
}

func (sel *instrumentsSelector) Instruments() []Instrument {
	instruments := make([]Instrument, len(sel.ids))
	for i, id := range sel.ids {
		instruments[i] = sel.instruments[id]
	}
	return instruments
}
