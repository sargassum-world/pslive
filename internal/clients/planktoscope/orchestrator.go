package planktoscope

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"golang.org/x/sync/errgroup"
)

type Orchestrator struct {
	planktoscopes   map[int64]*Client
	planktoscopesMu *sync.RWMutex

	logger godest.Logger
}

func NewOrchestrator(logger godest.Logger) *Orchestrator {
	return &Orchestrator{
		planktoscopes:   make(map[int64]*Client),
		planktoscopesMu: &sync.RWMutex{},
		logger:          logger,
	}
}

func (o *Orchestrator) Add(id int64, url string) error {
	if _, ok := o.Get(id); ok {
		return nil
	}

	config, err := GetConfig(url)
	if err != nil {
		return errors.Wrap(err, "couldn't set up planktoscope config")
	}
	client, err := NewClient(config, o.logger)
	if err != nil {
		return errors.Wrapf(err, "couldn't set up planktoscope client %d (%s)", id, url)
	}

	o.planktoscopesMu.Lock()
	defer o.planktoscopesMu.Unlock()

	o.planktoscopes[id] = client
	go func() {
		// FIXME: does this goroutine get leaked when a connection cannot be established? Or does
		// Disconnect cancel the Connect method's infinite loop?
		o.logger.Infof("adding planktoscope client %d (%s)", id, url)
		if err := client.Connect(); err != nil {
			o.logger.Error(errors.Wrapf(err, "couldn't add planktoscope client %d (%s)", id, url))
		}
	}()
	return nil
}

func (o *Orchestrator) Get(id int64) (c *Client, ok bool) {
	o.planktoscopesMu.RLock()
	defer o.planktoscopesMu.RUnlock()

	c, ok = o.planktoscopes[id]
	return c, ok
}

func (o *Orchestrator) Remove(ctx context.Context, id int64) error {
	o.planktoscopesMu.Lock()
	defer o.planktoscopesMu.Unlock()

	client, ok := o.planktoscopes[id]
	if !ok {
		return nil
	}
	o.logger.Infof("removing planktoscope client %d (%s)", id, client.Config.URL)
	err := client.Shutdown(ctx)
	if err != nil {
		client.Close()
	}
	delete(o.planktoscopes, id)
	return err
}

func (o *Orchestrator) Update(ctx context.Context, id int64, url string) error {
	o.planktoscopesMu.RLock()
	client, ok := o.planktoscopes[id]
	o.planktoscopesMu.RUnlock()
	if !ok {
		return o.Add(id, url)
	}

	if client.Config.URL == url {
		return nil
	}

	if err := o.Remove(ctx, id); err != nil {
		return errors.Wrapf(err, "couldn't remove old planktoscope client %d to update it", id)
	}
	return errors.Wrapf(o.Add(id, url), "couldn't add new planktoscope client %d to update it", id)
}

func (o *Orchestrator) Close(ctx context.Context) error {
	o.planktoscopesMu.Lock()
	defer o.planktoscopesMu.Unlock()

	eg, _ := errgroup.WithContext(ctx)
	for _, client := range o.planktoscopes {
		eg.Go(func(c *Client) func() error {
			return func() error {
				// We pass the parent context to isolate failure of one client's graceful shutdown from the
				// other clients' graceful shutdowns
				err := c.Shutdown(ctx)
				if err != nil {
					c.Close()
				}
				return err
			}
		}(client))
	}
	o.planktoscopes = nil
	return eg.Wait()
}
