package instruments

// Job Specification

type Schedule struct {
	Interval string `hcl:"interval"`
	Start    string `hcl:"start,optional"`
}

type Action struct {
	Target  ControllerID `hcl:"target"`
	Command string       `hcl:"command"`
}

type Specification struct {
	Schedule Schedule `hcl:"schedule,block"`
	Tags     []string `hcl:"tags"`
	Action   Action   `hcl:"action,block"`
}

// Automation Job

type ParsedJob struct {
	Name             string
	Type             string
	RawSpecification string
	Specification    Specification
}
