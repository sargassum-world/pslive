package instruments

// All Instruments

func (c *Client) GetInstruments() ([]Instrument, error) {
	// TODO: look up the controllers from a database, if one is specified!
	controllers := make([]Instrument, 0)
	envInstrument := c.Config.Instrument

	if envInstrument != (Instrument{}) {
		controllers = append(controllers, envInstrument)
	}
	return controllers, nil
}

// Individual Instrument

func (c *Client) FindInstrument(name string) (*Instrument, error) {
	controllers, err := c.GetInstruments()
	if err != nil {
		return nil, err
	}

	for _, v := range controllers {
		if v.Name == name {
			return &v, nil
		}
	}
	return nil, nil
}
