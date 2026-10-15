package climate

import (
	"errors"
	"fmt"
	"strings"
)

type Controller struct {
	repo        *Repository
	drv         Driver
	vent        VentPolicy
	minPressure float64
}

func NewController(repo *Repository, drv Driver, vent VentPolicy, minPressure float64) *Controller {
	return &Controller{repo: repo, drv: drv, vent: vent, minPressure: minPressure}
}

func (c *Controller) Start(ref, zone string, liters float64) (*Run, error) {
	if strings.TrimSpace(ref) == "" || strings.TrimSpace(zone) == "" || liters <= 0 {
		return nil, fmt.Errorf("%w: ref, zone and positive liters are required", ErrInvalidInput)
	}
	run := &Run{Ref: ref, Zone: zone, Liters: liters, VentLevel: c.vent.Level(liters), Status: RunPending}
	if err := c.repo.SaveRun(run); err != nil {
		return nil, err
	}
	if p := c.drv.Pressure(); p < c.minPressure {
		return c.failStart(run, fmt.Errorf("%w: %.1f bar < %.1f bar", ErrLowPressure, p, c.minPressure))
	}
	if err := c.drv.OpenValve(zone, liters); err != nil {
		return c.failStart(run, err)
	}
	if err := c.drv.SetVent(run.VentLevel); err != nil {
		_ = c.drv.CloseValve(zone)
		return c.failStart(run, err)
	}
	if err := run.Start(); err != nil {
		return run, err
	}
	return run, c.repo.SaveRun(run)
}

func (c *Controller) failStart(run *Run, cause error) (*Run, error) {
	_ = run.Abort(cause.Error())
	if err := c.repo.SaveRun(run); err != nil {
		return run, errors.Join(cause, err)
	}
	return run, cause
}

func (c *Controller) Finish(id string) error {
	run, ok := c.repo.GetRun(id)
	if !ok {
		return ErrRunNotFound
	}
	if err := run.Finish(); err != nil {
		return err
	}
	if err := c.drv.CloseValve(run.Zone); err != nil {
		return err
	}
	return c.repo.SaveRun(run)
}

func (c *Controller) Abort(id, reason string) error {
	run, ok := c.repo.GetRun(id)
	if !ok {
		return ErrRunNotFound
	}
	if err := run.Abort(reason); err != nil {
		return err
	}
	if err := c.drv.CloseValve(run.Zone); err != nil {
		return err
	}
	return c.repo.SaveRun(run)
}

func (c *Controller) FindByRef(ref string) (*Run, bool) { return c.repo.FindRunByRef(ref) }
func (c *Controller) Runs() []Run                       { return c.repo.Runs() }
