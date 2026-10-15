package soil_test

import (
	"errors"
	"path/filepath"
	"testing"

	"greenhouse/internal/soil"
)

func newScheduler(t *testing.T) (*soil.Scheduler, *soil.SensorArray, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "soil.json")
	repo, err := soil.OpenRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	sensors := soil.NewSensorArray(repo)
	return soil.NewScheduler(repo, sensors, soil.DefaultRule()), sensors, path
}

func TestReserveThenComplete(t *testing.T) {
	s, sensors, _ := newScheduler(t)
	_ = sensors.Record("z1", 20)
	p, err := s.Reserve("ref-1", "z1")
	if err != nil || p.Liters != 80 || p.Status != soil.PlanReserved {
		t.Fatalf("plan=%+v err=%v", p, err)
	}
	if err := s.Complete(p.ID); err != nil {
		t.Fatal(err)
	}
	if m, _ := sensors.Moisture("z1"); m != 60 {
		t.Fatalf("moisture after watering = %v, want 60", m)
	}
	if err := s.Complete(p.ID); !errors.Is(err, soil.ErrIllegalTransition) {
		t.Fatalf("second Complete: %v", err)
	}
}

func TestReserveRejections(t *testing.T) {
	s, sensors, _ := newScheduler(t)
	_ = sensors.Record("wet", 80)
	cases := []struct {
		name, ref, zone string
		want            error
	}{
		{"unknown zone", "r", "nope", soil.ErrUnknownZone},
		{"wet soil", "r", "wet", soil.ErrWateringNotNeeded},
		{"blank zone", "r", " ", soil.ErrInvalidInput},
		{"blank ref", "", "wet", soil.ErrInvalidInput},
	}
	for _, c := range cases {
		if _, err := s.Reserve(c.ref, c.zone); !errors.Is(err, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, err, c.want)
		}
	}
	if len(s.Plans()) != 0 {
		t.Fatal("a rejected reservation must not be persisted")
	}
	if err := sensors.Record("z", 101); !errors.Is(err, soil.ErrInvalidInput) {
		t.Fatalf("moisture 101: %v", err)
	}
}

func TestReleaseAndNotFound(t *testing.T) {
	s, sensors, _ := newScheduler(t)
	_ = sensors.Record("z1", 20)
	p, _ := s.Reserve("ref-1", "z1")
	if err := s.Release(p.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Release(p.ID); !errors.Is(err, soil.ErrIllegalTransition) {
		t.Fatalf("double release: %v", err)
	}
	if err := s.Release("nope"); !errors.Is(err, soil.ErrPlanNotFound) {
		t.Fatalf("release unknown: %v", err)
	}
	if err := s.Complete("nope"); !errors.Is(err, soil.ErrPlanNotFound) {
		t.Fatalf("complete unknown: %v", err)
	}
}

func TestStateSurvivesReopen(t *testing.T) {
	s, sensors, path := newScheduler(t)
	_ = sensors.Record("z1", 20)
	_, _ = s.Reserve("ref-1", "z1")

	repo2, err := soil.OpenRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	s2 := soil.NewScheduler(repo2, soil.NewSensorArray(repo2), soil.DefaultRule())
	if p, ok := s2.FindByRef("ref-1"); !ok || p.Status != soil.PlanReserved {
		t.Fatalf("plan lost after reopen: %+v", p)
	}
	if m, err := soil.NewSensorArray(repo2).Moisture("z1"); err != nil || m != 20 {
		t.Fatalf("reading lost: %v %v", m, err)
	}
}

func TestMoistureRule(t *testing.T) {
	r := soil.DefaultRule()
	if !r.NeedsWater(34) || r.NeedsWater(35) || r.LitersNeeded(70) != 0 {
		t.Fatal("rule thresholds are wrong")
	}
}
