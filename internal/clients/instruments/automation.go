package instruments

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
)

// Job Specification

func ParseSpecification(name, raw string) (parsed ParsedSpecification, err error) {
	output := &ParsedSpecification{}
	if err = hclsimple.Decode(name+".hcl", []byte(raw), nil, output); err != nil {
		return ParsedSpecification{}, err
	}
	return *output, nil
}

// Orchestrated Job

type OrchestratedJob struct {
	ID           AutomationJobID
	InstrumentID InstrumentID
	Name         string
	Type         string
	RawSpec      string
	ParsedSpec   ParsedSpecification
	startedJob   *gocron.Job
	canceler     func()
}

func NewOrchestratedJob(
	id AutomationJobID, instrumentID InstrumentID, name, specType, rawSpec string,
) (job *OrchestratedJob, err error) {
	job = &OrchestratedJob{
		ID:           id,
		InstrumentID: instrumentID,
		Name:         name,
		Type:         specType,
		RawSpec:      rawSpec,
	}
	switch specType {
	default:
		return nil, errors.Errorf("unknown specification type %s", specType)
	case "hcl-v0.1.0":
		if job.ParsedSpec, err = ParseSpecification(name, rawSpec); err != nil {
			return nil, errors.Wrapf(err, "couldn't parse %s specification", specType)
		}
	}
	return job, nil
}

func (j *OrchestratedJob) Run(ctx context.Context) error {
	// TODO: implement
	fmt.Printf("%+v\n", j.ParsedSpec.Action)
	return nil
}

func (j *OrchestratedJob) Cancel() {
	if j.canceler == nil {
		return
	}
	j.canceler()
}

// Job Orchestrator

type JobOrchestrator struct {
	jobs      map[AutomationJobID]*OrchestratedJob
	jobsMu    *sync.RWMutex
	scheduler *gocron.Scheduler
	toStart   chan *OrchestratedJob

	logger godest.Logger
}

func NewJobOrchestrator(logger godest.Logger) *JobOrchestrator {
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.StartAsync()
	scheduler.SingletonModeAll()
	return &JobOrchestrator{
		jobs:      make(map[AutomationJobID]*OrchestratedJob),
		jobsMu:    &sync.RWMutex{},
		scheduler: scheduler,
		toStart:   make(chan *OrchestratedJob),
		logger:    logger,
	}
}

func (o *JobOrchestrator) startJob(ctx context.Context, job *OrchestratedJob) error {
	schedule := job.ParsedSpec.Schedule
	startTime, err := schedule.DecodeStart()
	if err != nil {
		return err
	}
	o.scheduler.Every(schedule.Interval)
	if startTime != nil {
		o.scheduler.StartAt(*startTime)
	}

	jobCtx, canceler := context.WithCancel(ctx)
	job.canceler = canceler
	job.startedJob, err = o.scheduler.Do(func() {
		select {
		case <-jobCtx.Done():
			return
		default:
			if ctxErr := jobCtx.Err(); ctxErr != nil {
				// TODO: log any relevant errors?
				return
			}

			if jobErr := job.Run(jobCtx); jobErr != nil {
				o.logger.Error(errors.Wrapf(jobErr, "job %d %s failed", job.ID, job.Name))
			}
		}
	})
	return errors.Wrapf(err, "couldn't start job %d %s", job.ID, job.Name)
}

func (o *JobOrchestrator) Orchestrate(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case job := <-o.toStart:
			if err := ctx.Err(); err != nil {
				// Context was also canceled and it should have priority
				return err
			}

			o.logger.Infof("starting job %d %s", job.ID, job.Name)
			if err := o.startJob(ctx, job); err != nil {
				o.logger.Error(err)
			}
		}
	}
}

func (o *JobOrchestrator) Add(
	id AutomationJobID, instrumentID InstrumentID, name, specType, rawSpec string,
) error {
	if _, ok := o.Get(id); ok {
		o.logger.Warnf("skipped adding job %d %s because it's already running", id, name)
		return nil
	}

	if name == "" {
		name = fmt.Sprint(id)
	}
	job, err := NewOrchestratedJob(id, instrumentID, name, specType, rawSpec)
	if err != nil {
		// TODO: if there's an HCL parsing error, we should report diagnostics in the GUI
		return errors.Wrapf(
			err, "couldn't create job %d %s", id, name,
		)
	}

	o.jobsMu.Lock()
	o.jobs[id] = job
	o.jobsMu.Unlock()

	o.logger.Debugf("adding job %d %s to start queue", id, name)
	o.toStart <- job
	return nil
}

func (o *JobOrchestrator) Get(id AutomationJobID) (j *OrchestratedJob, ok bool) {
	o.jobsMu.RLock()
	defer o.jobsMu.RUnlock()

	j, ok = o.jobs[id]
	return j, ok
}

func (o *JobOrchestrator) Remove(id AutomationJobID) {
	o.jobsMu.Lock()
	defer o.jobsMu.Unlock()

	job, ok := o.jobs[id]
	if !ok {
		return
	}
	o.logger.Debugf("removing job %d", id, job.Name)
	job.Cancel()
	o.scheduler.RemoveByReference(job.startedJob)
	delete(o.jobs, id)
	o.logger.Infof("removed job %d %s", id, job.Name)
}

func (o *JobOrchestrator) Update(
	id AutomationJobID, instrumentID InstrumentID, name, specType, rawSpec string,
) error {
	o.jobsMu.RLock()
	_, ok := o.jobs[id]
	o.jobsMu.RUnlock()
	if name == "" {
		name = fmt.Sprint(id)
	}
	if !ok {
		return o.Add(id, instrumentID, name, specType, rawSpec)
	}

	o.Remove(id)
	return errors.Wrapf(
		o.Add(id, instrumentID, name, specType, rawSpec),
		"couldn't add new job %d %s to update it", id, name,
	)
}

func (o *JobOrchestrator) Close() {
	o.jobsMu.Lock()
	defer o.jobsMu.Unlock()

	for _, job := range o.jobs {
		job.Cancel()
		o.scheduler.RemoveByReference(job.startedJob)
	}
	o.scheduler.Stop()

	if o.toStart != nil {
		close(o.toStart)
		o.toStart = nil
	}
}
