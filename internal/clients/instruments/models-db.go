package instruments

import (
	"zombiezen.com/go/sqlite"
)

type (
	InstrumentID    int64
	AdminID         string
	CameraID        int64
	ControllerID    int64
	AutomationJobID int64
)

// Camera

type Camera struct {
	ID           CameraID
	InstrumentID InstrumentID
	Enabled      bool
	Name         string
	Description  string
	Protocol     string
	URL          string
}

func (c Camera) addParams(params map[string]interface{}) (fullParams map[string]interface{}) {
	fullParams = make(map[string]interface{})
	for key, value := range params {
		fullParams[key] = value
	}
	for key, value := range map[string]interface{}{
		"$enabled":     c.Enabled,
		"$name":        c.Name,
		"$description": c.Description,
		"$protocol":    c.Protocol,
		"$url":         c.URL,
	} {
		fullParams[key] = value
	}
	return fullParams
}

func (c Camera) newInsertion() map[string]interface{} {
	return c.addParams(map[string]interface{}{"$instrument_id": c.InstrumentID})
}

func (c Camera) newUpdate() map[string]interface{} {
	return c.addParams(map[string]interface{}{"$id": c.ID})
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

func getCamera(
	s *sqlite.Stmt, fieldPrefix string, id CameraID, instrumentID InstrumentID,
) Camera {
	return Camera{
		ID:           id,
		InstrumentID: instrumentID,
		Enabled:      s.GetBool(fieldPrefix + "enabled"),
		Name:         s.GetText(fieldPrefix + "name"),
		Description:  s.GetText(fieldPrefix + "description"),
		Protocol:     s.GetText(fieldPrefix + "protocol"),
		URL:          s.GetText(fieldPrefix + "url"),
	}
}

func (sel *camerasSelector) Step(s *sqlite.Stmt) error {
	id := CameraID(s.GetInt64("id"))
	if _, ok := sel.cameras[id]; !ok {
		sel.cameras[id] = getCamera(s, "", id, InstrumentID(s.GetInt64("instrument_id")))
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
	Enabled      bool
	Name         string
	Description  string
	Protocol     string
	URL          string
}

func (c Controller) addParams(params map[string]interface{}) (fullParams map[string]interface{}) {
	fullParams = make(map[string]interface{})
	for key, value := range params {
		fullParams[key] = value
	}
	for key, value := range map[string]interface{}{
		"$enabled":     c.Enabled,
		"$name":        c.Name,
		"$description": c.Description,
		"$protocol":    c.Protocol,
		"$url":         c.URL,
	} {
		fullParams[key] = value
	}
	return fullParams
}

func (c Controller) newInsertion() map[string]interface{} {
	return c.addParams(map[string]interface{}{"$instrument_id": c.InstrumentID})
}

func (c Controller) newUpdate() map[string]interface{} {
	return c.addParams(map[string]interface{}{"$id": c.ID})
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

func getController(
	s *sqlite.Stmt, fieldPrefix string, id ControllerID, instrumentID InstrumentID,
) Controller {
	return Controller{
		ID:           id,
		InstrumentID: instrumentID,
		Enabled:      s.GetBool(fieldPrefix + "enabled"),
		Name:         s.GetText(fieldPrefix + "name"),
		Description:  s.GetText(fieldPrefix + "description"),
		Protocol:     s.GetText(fieldPrefix + "protocol"),
		URL:          s.GetText(fieldPrefix + "url"),
	}
}

func (sel *controllersSelector) Step(s *sqlite.Stmt) error {
	id := ControllerID(s.GetInt64("id"))
	if _, ok := sel.controllers[id]; !ok {
		sel.controllers[id] = getController(s, "", id, InstrumentID(s.GetInt64("instrument_id")))
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

// Automation Job

type AutomationJob struct {
	ID            AutomationJobID
	InstrumentID  InstrumentID
	Enabled       bool
	Name          string
	Description   string
	Type          string
	Specification string
}

func (j AutomationJob) addParams(
	params map[string]interface{},
) (fullParams map[string]interface{}) {
	fullParams = make(map[string]interface{})
	for key, value := range params {
		fullParams[key] = value
	}
	for key, value := range map[string]interface{}{
		"$enabled":       j.Enabled,
		"$name":          j.Name,
		"$description":   j.Description,
		"$type":          j.Type,
		"$specification": j.Specification,
	} {
		fullParams[key] = value
	}
	return fullParams
}

func (j AutomationJob) newInsertion() map[string]interface{} {
	return j.addParams(map[string]interface{}{"$instrument_id": j.InstrumentID})
}

func (j AutomationJob) newUpdate() map[string]interface{} {
	return j.addParams(map[string]interface{}{"$id": j.ID})
}

func (j AutomationJob) newDelete() map[string]interface{} {
	return map[string]interface{}{
		"$id": j.ID,
	}
}

// Automation Jobs

type automationJobsSelector struct {
	ids            []AutomationJobID
	automationJobs map[AutomationJobID]AutomationJob
}

func newAutomationJobsSelector() *automationJobsSelector {
	return &automationJobsSelector{
		ids:            make([]AutomationJobID, 0),
		automationJobs: make(map[AutomationJobID]AutomationJob),
	}
}

func getAutomationJob(
	s *sqlite.Stmt, fieldPrefix string, id AutomationJobID, instrumentID InstrumentID,
) AutomationJob {
	return AutomationJob{
		ID:            id,
		InstrumentID:  instrumentID,
		Enabled:       s.GetBool(fieldPrefix + "enabled"),
		Name:          s.GetText(fieldPrefix + "name"),
		Description:   s.GetText(fieldPrefix + "description"),
		Type:          s.GetText(fieldPrefix + "type"),
		Specification: s.GetText(fieldPrefix + "specification"),
	}
}

func (sel *automationJobsSelector) Step(s *sqlite.Stmt) error {
	id := AutomationJobID(s.GetInt64("id"))
	if _, ok := sel.automationJobs[id]; !ok {
		sel.automationJobs[id] = getAutomationJob(s, "", id, InstrumentID(s.GetInt64("instrument_id")))
		if id != 0 {
			sel.ids = append(sel.ids, id)
		}
	}
	return nil
}

func (sel *automationJobsSelector) AutomationJobs() []AutomationJob {
	automationJobs := make([]AutomationJob, len(sel.ids))
	for i, id := range sel.ids {
		automationJobs[i] = sel.automationJobs[id]
	}
	return automationJobs
}

// Instrument

type Instrument struct {
	ID             InstrumentID
	Name           string
	Description    string
	AdminID        AdminID
	Cameras        map[CameraID]Camera
	Controllers    map[ControllerID]Controller
	AutomationJobs map[AutomationJobID]AutomationJob
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
			ID:             instrumentID,
			Name:           s.GetText("name"),
			Description:    s.GetText("description"),
			AdminID:        AdminID(s.GetText("admin_id")),
			Cameras:        make(map[CameraID]Camera),
			Controllers:    make(map[ControllerID]Controller),
			AutomationJobs: make(map[AutomationJobID]AutomationJob),
		}
		if instrumentID != 0 {
			sel.ids = append(sel.ids, instrumentID)
		}
	}
	instrument := sel.instruments[instrumentID]

	cameraID := CameraID(s.GetInt64("camera_id"))
	if camera := getCamera(s, "camera_", cameraID, instrumentID); camera != (Camera{
		ID:           cameraID,
		InstrumentID: instrumentID,
	}) {
		instrument.Cameras[cameraID] = camera
	}

	controllerID := ControllerID(s.GetInt64("controller_id"))
	if controller := getController(
		s, "controller_", controllerID, instrumentID,
	); controller != (Controller{
		ID:           controllerID,
		InstrumentID: instrumentID,
	}) {
		instrument.Controllers[controllerID] = controller
	}

	automationJobID := AutomationJobID(s.GetInt64("automation_job_id"))
	if automationJob := getAutomationJob(
		s, "automation_job_", automationJobID, instrumentID,
	); automationJob != (AutomationJob{
		ID:           automationJobID,
		InstrumentID: instrumentID,
	}) {
		instrument.AutomationJobs[automationJobID] = automationJob
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
