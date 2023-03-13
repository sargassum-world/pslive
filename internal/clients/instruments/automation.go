package instruments

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"golang.org/x/sync/errgroup"
)

// Automation Job

type ParsedJob struct {
	Type          string
	Specification string
}

func NewParsedJob(
	specificationType, specification string, logger godest.Logger,
) (job *ParsedJob, err error) {
	// TODO: actually parse the job here
	return &ParsedJob{
		Type:          specificationType,
		Specification: specification,
	}, nil
}

func (j *ParsedJob) Start() error {
	return nil
}

func (j *ParsedJob) Shutdown(ctx context.Context) error {
	// TODO: if possible, use context for cancellations, rather than a separate Shutdown or Close
	// method
	return nil
}

func (j *ParsedJob) Close() {
}

// Automation Job Orchestrator

type AutomationJobOrchestrator struct {
	jobs   map[AutomationJobID]*ParsedJob
	jobsMu *sync.RWMutex

	logger godest.Logger
}

func NewAutomationJobOrchestrator(logger godest.Logger) *AutomationJobOrchestrator {
	return &AutomationJobOrchestrator{
		jobs:   make(map[AutomationJobID]*ParsedJob),
		jobsMu: &sync.RWMutex{},
		logger: logger,
	}
}

func (o *AutomationJobOrchestrator) Add(
	id AutomationJobID, specificationType, specification string,
) error {
	if _, ok := o.Get(id); ok {
		o.logger.Warnf("skipped adding automation job %d because it's already running", id)
		return nil
	}

	job, err := NewParsedJob(specificationType, specification, o.logger)
	if err != nil {
		return errors.Wrapf(
			err, "couldn't create automation job %d from %s specification", id, specificationType,
		)
	}

	o.jobsMu.Lock()
	o.jobs[id] = job
	o.jobsMu.Unlock()

	go func() {
		o.logger.Infof("starting automation job %d", id)
		if err := job.Start(); err != nil {
			o.logger.Error(errors.Wrapf(err, "couldn't starting automation job %d", id))
		}
	}()
	return nil
}

func (o *AutomationJobOrchestrator) Get(id AutomationJobID) (c *ParsedJob, ok bool) {
	o.jobsMu.RLock()
	defer o.jobsMu.RUnlock()

	c, ok = o.jobs[id]
	return c, ok
}

func (o *AutomationJobOrchestrator) Remove(ctx context.Context, id AutomationJobID) error {
	o.jobsMu.Lock()
	defer o.jobsMu.Unlock()

	job, ok := o.jobs[id]
	if !ok {
		return nil
	}
	o.logger.Infof("removing automation job %d", id)
	err := job.Shutdown(ctx)
	if err != nil {
		job.Close()
	}
	delete(o.jobs, id)
	return err
}

func (o *AutomationJobOrchestrator) Update(
	ctx context.Context, id AutomationJobID, specificationType, specification string,
) error {
	o.jobsMu.RLock()
	_, ok := o.jobs[id]
	o.jobsMu.RUnlock()
	if !ok {
		return o.Add(id, specificationType, specification)
	}

	if err := o.Remove(ctx, id); err != nil {
		return errors.Wrapf(err, "couldn't remove old automation job %d to update it", id)
	}
	return errors.Wrapf(
		o.Add(id, specificationType, specification),
		"couldn't add new automation job %d to update it", id,
	)
}

func (o *AutomationJobOrchestrator) Close(ctx context.Context) error {
	o.jobsMu.Lock()
	defer o.jobsMu.Unlock()

	eg, _ := errgroup.WithContext(ctx)
	for _, job := range o.jobs {
		eg.Go(func(c *ParsedJob) func() error {
			return func() error {
				// We pass the parent context to isolate failure of one job's graceful shutdown from the
				// other jobs' graceful shutdowns
				err := c.Shutdown(ctx)
				if err != nil {
					c.Close()
				}
				return err
			}
		}(job))
	}
	o.jobs = nil
	return eg.Wait()
}
