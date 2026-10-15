package facade

import (
	"fmt"

	"greenhouse/internal/climate"
	"greenhouse/internal/dosing"
	"greenhouse/internal/soil"
)

type AdminFacade struct{ *core }

type Snapshot struct {
	Chains   []Chain
	Plans    []soil.WateringPlan
	Runs     []climate.Run
	Doses    []dosing.Dose
	Moisture map[string]float64
	TankMl   float64
}

func (a *AdminFacade) Snapshot() Snapshot {
	return Snapshot{
		Chains:   a.chains.All(),
		Plans:    a.sched.Plans(),
		Runs:     a.ctrl.Runs(),
		Doses:    a.disp.Doses(),
		Moisture: a.sensors.All(),
		TankMl:   a.disp.TankLevel(),
	}
}

func (a *AdminFacade) Recover() ([]Chain, error) {
	var done []Chain
	for _, ch := range a.chains.All() {
		if !ch.Unfinished() {
			continue
		}
		c := ch
		if err := a.compensate(&c); err != nil {
			return done, fmt.Errorf("recover %s: %w", c.ID, err)
		}
		done = append(done, c)
	}
	return done, nil
}

func (a *AdminFacade) Compensate(id string) (*Chain, error) {
	ch, ok := a.chains.Get(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrChainNotFound, id)
	}
	if ch.Status == Completed || ch.Status == Compensated {
		return ch, fmt.Errorf("%w: %s is %s", ErrNotCompensable, id, ch.Status)
	}
	return ch, a.compensate(ch)
}

func (a *AdminFacade) RecordMoisture(zone string, pct float64) error {
	return a.sensors.Record(zone, pct)
}

func (a *AdminFacade) SetTankLevel(ml float64) error { return a.disp.SetTankLevel(ml) }
