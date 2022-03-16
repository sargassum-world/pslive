package instruments

type Instrument struct {
	MJPEGStream string `json:"mjpegStream"`
	Controller  string `json:"controller"`
	Name        string `json:"name"` // Must be unique for display purposes!
	Description string `json:"description"`
}
