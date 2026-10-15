package facade

import "fmt"

type GrowthFacade struct {
	*core

	AfterB func() error
}

func (f *GrowthFacade) WaterZone(zone string) (*Chain, error) {
	id, err := f.chains.NewID()
	if err != nil {
		return nil, err
	}

	plan, err := f.sched.Reserve(id, zone)
	if err != nil {
		return nil, err
	}
	ch := &Chain{ID: id, Zone: zone, Liters: plan.Liters, Status: PendingB}
	if err := f.chains.Save(ch); err != nil {
		return ch, err
	}

	run, err := f.ctrl.Start(id, zone, plan.Liters)
	if err != nil {
		return f.fail(ch, "B", err)
	}
	ch.Status = PendingC
	if err := f.chains.Save(ch); err != nil {
		return ch, err
	}

	if f.AfterB != nil {
		if err := f.AfterB(); err != nil {
			return ch, err
		}
	}

	if _, err := f.disp.Dispense(id, zone, plan.Liters); err != nil {
		return f.fail(ch, "C", err)
	}

	if err := f.ctrl.Finish(run.ID); err != nil {
		return ch, err
	}
	if err := f.sched.Complete(plan.ID); err != nil {
		return ch, err
	}
	ch.Status = Completed
	return ch, f.chains.Save(ch)
}

func (f *GrowthFacade) fail(ch *Chain, step string, cause error) (*Chain, error) {
	ch.Error = cause.Error()
	if err := f.compensate(ch); err != nil {
		return ch, fmt.Errorf("compensation failed, chain %s left in %s: %w", ch.ID, ch.Status, err)
	}
	return ch, &StepError{Step: step, Err: cause, Compensated: ch.Compensated}
}
