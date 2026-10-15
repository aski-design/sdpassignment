package soil

import (
	"fmt"
	"strings"
)

type Scheduler struct {
	repo    *Repository
	sensors *SensorArray
	rule    MoistureRule
}

func NewScheduler(repo *Repository, sensors *SensorArray, rule MoistureRule) *Scheduler {
	return &Scheduler{repo: repo, sensors: sensors, rule: rule}
}

func (s *Scheduler) Reserve(ref, zone string) (*WateringPlan, error) {
	if strings.TrimSpace(ref) == "" || strings.TrimSpace(zone) == "" {
		return nil, fmt.Errorf("%w: ref and zone are required", ErrInvalidInput)
	}
	m, err := s.sensors.Moisture(zone)
	if err != nil {
		return nil, err
	}
	if !s.rule.NeedsWater(m) {
		return nil, fmt.Errorf("%w: %s is at %.0f%%", ErrWateringNotNeeded, zone, m)
	}
	p := &WateringPlan{Ref: ref, Zone: zone, Liters: s.rule.LitersNeeded(m), Status: PlanReserved}
	if err := s.repo.SavePlan(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Scheduler) Complete(id string) error {
	p, ok := s.repo.GetPlan(id)
	if !ok {
		return ErrPlanNotFound
	}
	if err := p.Complete(); err != nil {
		return err
	}
	if err := s.repo.SavePlan(p); err != nil {
		return err
	}
	return s.sensors.Record(p.Zone, s.rule.Target)
}

func (s *Scheduler) Release(id string) error {
	p, ok := s.repo.GetPlan(id)
	if !ok {
		return ErrPlanNotFound
	}
	if err := p.Release(); err != nil {
		return err
	}
	return s.repo.SavePlan(p)
}

func (s *Scheduler) FindByRef(ref string) (*WateringPlan, bool) { return s.repo.FindPlanByRef(ref) }
func (s *Scheduler) Plans() []WateringPlan                      { return s.repo.Plans() }
