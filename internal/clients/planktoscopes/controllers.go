package planktoscopes

// All Planktoscopes

func (c *Client) GetPlanktoscopes() ([]Planktoscope, error) {
	// TODO: look up the controllers from a database, if one is specified!
	controllers := make([]Planktoscope, 0)
	envPlanktoscope := c.Config.Planktoscope

	if envPlanktoscope != (Planktoscope{}) {
		controllers = append(controllers, envPlanktoscope)
	}
	return controllers, nil
}

// Individual Planktoscope

func (c *Client) FindPlanktoscope(name string) (*Planktoscope, error) {
	controllers, err := c.GetPlanktoscopes()
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
