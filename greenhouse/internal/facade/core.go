package facade

import (
	"greenhouse/internal/climate"
	"greenhouse/internal/dosing"
	"greenhouse/internal/soil"
)

type core struct {
	sensors *soil.SensorArray
	sched   *soil.Scheduler
	ctrl    *climate.Controller
	disp    *dosing.Dispenser
	chains  *ChainRepo
}

func (c *core) compensate(ch *Chain) error {
	ch.Status = Compensating
	ch.Compensated = nil
	if err := c.chains.Save(ch); err != nil {
		return err
	}
	if run, ok := c.ctrl.FindByRef(ch.ID); ok && run.Active() {
		if err := c.ctrl.Abort(run.ID, "compensation of "+ch.ID); err != nil {
			return err
		}
		ch.Compensated = append(ch.Compensated, "B")
	}
	if plan, ok := c.sched.FindByRef(ch.ID); ok && plan.Status == soil.PlanReserved {
		if err := c.sched.Release(plan.ID); err != nil {
			return err
		}
		ch.Compensated = append(ch.Compensated, "A")
	}
	ch.Status = Compensated
	return c.chains.Save(ch)
}
