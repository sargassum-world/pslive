package planktoscopes

type Planktoscope struct {
	MJPEGStream string `json:"mjpegStream"`
	Name        string `json:"name"` // Must be unique for display purposes!
	Description string `json:"description"`
}
