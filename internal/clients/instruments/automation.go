package instruments

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
)

// Job Specification

func parseSpecification(name, raw string) (parsed ParsedSpecification, err error) {
	output := &ParsedSpecification{}
	if err = hclsimple.Decode(name+".hcl", []byte(raw), nil, output); err != nil {
		return ParsedSpecification{}, err
	}
	return *output, nil
}

// Job Actions

type ActionHandler func(
	ctx context.Context, instrumentID InstrumentID, name string, params hcl.Body,
) error

func sleep(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func HandleSleepAction(ctx context.Context, _ InstrumentID, name string, params hcl.Body) error {
	var a SleepAction
	if err := gohcl.DecodeBody(params, nil, &a); err != nil {
		return errors.Wrapf(err, "couldn't decode sleep action %s", name)
	}
	duration, err := time.ParseDuration(a.Duration)
	if err != nil {
		return errors.Wrapf(err, "couldn't parse sleep duration %s", a.Duration)
	}

	if err := sleep(ctx, duration); err != nil {
		return errors.Wrapf(err, "couldn't run sleep action %s", name)
	}
	return nil
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
		if job.ParsedSpec, err = parseSpecification(name, rawSpec); err != nil {
			return nil, errors.Wrapf(err, "couldn't parse %s specification", specType)
		}
	}
	return job, nil
}

func (j *OrchestratedJob) Run(ctx context.Context, handlers map[string]ActionHandler) error {
	for i, action := range j.ParsedSpec.Actions {
		handler, ok := handlers[action.Type]
		if !ok {
			return errors.Errorf("action #%d (%s) has unhandled type %s", i, action.Name, action.Type)
		}
		if err := handler(ctx, j.InstrumentID, action.Name, action.Remain); err != nil {
			return errors.Wrapf(err, "handler for %s action #%d (%s) failed", action.Type, i, action.Name)
		}
	}
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
	jobs           map[AutomationJobID]*OrchestratedJob
	mu             *sync.RWMutex
	scheduler      *gocron.Scheduler
	toStart        chan *OrchestratedJob
	canceler       func()
	actionHandlers map[string]ActionHandler

	logger godest.Logger
}

func NewJobOrchestrator(
	actionHandlers map[string]ActionHandler, logger godest.Logger,
) *JobOrchestrator {
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.StartAsync()
	scheduler.SingletonModeAll()
	return &JobOrchestrator{
		mu:             &sync.RWMutex{},
		jobs:           make(map[AutomationJobID]*OrchestratedJob),
		scheduler:      scheduler,
		toStart:        make(chan *OrchestratedJob),
		actionHandlers: actionHandlers,
		logger:         logger,
	}
}

func (o *JobOrchestrator) startJob(ctx context.Context, job *OrchestratedJob) error {
	schedule := job.ParsedSpec.Schedule
	startTime, err := schedule.DecodeStart()
	if err != nil {
		return err
	}

	o.mu.RLock()
	defer o.mu.RUnlock()

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

			if jobErr := job.Run(jobCtx, o.actionHandlers); jobErr != nil {
				o.logger.Error(errors.Wrapf(jobErr, "job %d %s failed", job.ID, job.Name))
			}
		}
	})
	return errors.Wrapf(err, "couldn't start job %d %s", job.ID, job.Name)
}

func (o *JobOrchestrator) Orchestrate(ctx context.Context) error {
	o.mu.Lock()
	ctx, o.canceler = context.WithCancel(ctx)
	o.mu.Unlock()

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

	o.mu.Lock()
	o.jobs[id] = job
	o.mu.Unlock()

	o.logger.Debugf("adding job %d %s to start queue", id, name)
	o.toStart <- job
	return nil
}

func (o *JobOrchestrator) Get(id AutomationJobID) (j *OrchestratedJob, ok bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	j, ok = o.jobs[id]
	return j, ok
}

func (o *JobOrchestrator) Remove(id AutomationJobID) {
	o.mu.Lock()
	defer o.mu.Unlock()

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
	o.mu.RLock()
	_, ok := o.jobs[id]
	o.mu.RUnlock()
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
	o.mu.Lock()
	defer o.mu.Unlock()

	o.canceler()

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
