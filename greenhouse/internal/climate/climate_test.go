package climate_test

import (
	"errors"
	"path/filepath"
	"testing"

	"greenhouse/internal/climate"
)

func newController(t *testing.T) (*climate.Controller, *climate.SimDriver, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "climate.json")
	repo, err := climate.OpenRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	drv := climate.NewSimDriver()
	return climate.NewController(repo, drv, climate.DefaultVentPolicy(), 1.5), drv, path
}

func TestStartThenFinish(t *testing.T) {
	c, drv, _ := newController(t)
	run, err := c.Start("r1", "z1", 70)
	if err != nil || run.Status != climate.RunRunning || run.VentLevel != 100 {
		t.Fatalf("run=%+v err=%v", run, err)
	}
	if !drv.IsOpen("z1") || drv.VentLevel() != 100 {
		t.Fatal("valve should be open and vent at 100")
	}
	if err := c.Finish(run.ID); err != nil {
		t.Fatal(err)
	}
	if drv.IsOpen("z1") {
		t.Fatal("valve should be closed after Finish")
	}
}

func TestStartFailuresArePersistedAsAborted(t *testing.T) {
	c, drv, _ := newController(t)
	drv.SetPressure(0.5)
	if _, err := c.Start("r1", "z1", 10); !errors.Is(err, climate.ErrLowPressure) {
		t.Fatalf("low pressure: %v", err)
	}
	drv.SetPressure(3)
	drv.SetStuck("z2", true)
	if _, err := c.Start("r2", "z2", 10); !errors.Is(err, climate.ErrValveStuck) {
		t.Fatalf("stuck valve: %v", err)
	}
	for _, ref := range []string{"r1", "r2"} {
		if r, ok := c.FindByRef(ref); !ok || r.Status != climate.RunAborted || r.Reason == "" {
			t.Fatalf("%s: %+v", ref, r)
		}
	}
}

func TestAbortAndIllegalTransitions(t *testing.T) {
	c, drv, _ := newController(t)
	run, _ := c.Start("r1", "z1", 10)
	if err := c.Abort(run.ID, "test"); err != nil {
		t.Fatal(err)
	}
	if drv.IsOpen("z1") {
		t.Fatal("valve must be closed after Abort")
	}
	if err := c.Abort(run.ID, "again"); !errors.Is(err, climate.ErrIllegalTransition) {
		t.Fatalf("double abort: %v", err)
	}
	if err := c.Finish(run.ID); !errors.Is(err, climate.ErrIllegalTransition) {
		t.Fatalf("finish aborted: %v", err)
	}
	if err := c.Abort("nope", "x"); !errors.Is(err, climate.ErrRunNotFound) {
		t.Fatalf("abort unknown: %v", err)
	}
	if err := c.Finish("nope"); !errors.Is(err, climate.ErrRunNotFound) {
		t.Fatalf("finish unknown: %v", err)
	}
}

func TestInvalidInputAndPersistence(t *testing.T) {
	c, _, path := newController(t)
	if _, err := c.Start("", "z", 10); !errors.Is(err, climate.ErrInvalidInput) {
		t.Fatalf("blank ref: %v", err)
	}
	if _, err := c.Start("r", "z", 0); !errors.Is(err, climate.ErrInvalidInput) {
		t.Fatalf("zero liters: %v", err)
	}
	_, _ = c.Start("r1", "z1", 10)

	repo2, err := climate.OpenRepository(path)
	if err != nil {
		t.Fatal(err)
	}
	c2 := climate.NewController(repo2, climate.NewSimDriver(), climate.DefaultVentPolicy(), 1.5)
	if r, ok := c2.FindByRef("r1"); !ok || r.Status != climate.RunRunning || len(c2.Runs()) != 1 {
		t.Fatalf("run lost after reopen: %+v", r)
	}
}

func TestVentPolicy(t *testing.T) {
	p := climate.DefaultVentPolicy()
	if p.Level(1) != 20 || p.Level(20) != 30 || p.Level(100) != 100 {
		t.Fatal("vent levels are wrong")
	}
}
