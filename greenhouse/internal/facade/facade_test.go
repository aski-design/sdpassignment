package facade_test

import (
	"errors"
	"reflect"
	"testing"

	"greenhouse/internal/climate"
	"greenhouse/internal/dosing"
	"greenhouse/internal/facade"
	"greenhouse/internal/soil"
)

var errCrash = errors.New("simulated process crash")

type greenhouse struct {
	t      *testing.T
	dir    string
	drv    *climate.SimDriver
	growth *facade.GrowthFacade
	admin  *facade.AdminFacade
}

func aGreenhouse(t *testing.T) *greenhouse {
	t.Helper()
	g := &greenhouse{t: t, dir: t.TempDir()}
	g.boot()
	return g.withDrySoil("tomatoes")
}

func (g *greenhouse) boot() {
	g.t.Helper()
	g.drv = climate.NewSimDriver()
	growth, admin, err := facade.Assemble(g.dir, g.drv)
	if err != nil {
		g.t.Fatal(err)
	}
	g.growth, g.admin = growth, admin
}

func (g *greenhouse) must(err error) {
	g.t.Helper()
	if err != nil {
		g.t.Fatal(err)
	}
}

func (g *greenhouse) withDrySoil(zone string) *greenhouse {
	g.t.Helper()
	g.must(g.admin.RecordMoisture(zone, 25))
	return g
}
func (g *greenhouse) withWetSoil(zone string) *greenhouse {
	g.t.Helper()
	g.must(g.admin.RecordMoisture(zone, 80))
	return g
}
func (g *greenhouse) withLowWaterPressure() *greenhouse { g.drv.SetPressure(0.4); return g }
func (g *greenhouse) withStuckValve(zone string) *greenhouse {
	g.drv.SetStuck(zone, true)
	return g
}
func (g *greenhouse) withNutrientTank(ml float64) *greenhouse {
	g.t.Helper()
	g.must(g.admin.SetTankLevel(ml))
	return g
}
func (g *greenhouse) crashingAfterB() *greenhouse {
	g.growth.AfterB = func() error { return errCrash }
	return g
}

func (g *greenhouse) afterRestart() persisted {
	g.t.Helper()
	g.boot()
	return persisted{t: g.t, s: g.admin.Snapshot()}
}

func (g *greenhouse) waters(zone string) *outcome {
	ch, err := g.growth.WaterZone(zone)
	return &outcome{t: g.t, chain: ch, err: err}
}

type outcome struct {
	t     *testing.T
	chain *facade.Chain
	err   error
}

func (o *outcome) succeeds() *outcome {
	o.t.Helper()
	if o.err != nil {
		o.t.Fatalf("expected success, got %v", o.err)
	}
	return o
}

func (o *outcome) failsAtStep(step string) *outcome {
	o.t.Helper()
	var se *facade.StepError
	if !errors.As(o.err, &se) || se.Step != step {
		o.t.Fatalf("expected failure at step %s, got %v", step, o.err)
	}
	return o
}

func (o *outcome) because(target error) *outcome {
	o.t.Helper()
	if !errors.Is(o.err, target) {
		o.t.Fatalf("expected cause %v, got %v", target, o.err)
	}
	return o
}

func (o *outcome) compensated(steps ...string) *outcome {
	o.t.Helper()
	if !reflect.DeepEqual(o.chain.Compensated, steps) {
		o.t.Fatalf("compensated %v, want %v (order matters)", o.chain.Compensated, steps)
	}
	return o
}

func (o *outcome) isRejectedBecause(target error) *outcome {
	o.t.Helper()
	var se *facade.StepError
	if o.chain != nil || errors.As(o.err, &se) || !errors.Is(o.err, target) {
		o.t.Fatalf("expected a clean rejection because of %v, got chain=%v err=%v", target, o.chain, o.err)
	}
	return o
}

func (o *outcome) diesWith(target error) *outcome {
	o.t.Helper()
	if o.chain == nil || !errors.Is(o.err, target) {
		o.t.Fatalf("expected the process to die with %v, got %v", target, o.err)
	}
	return o
}

type persisted struct {
	t *testing.T
	s facade.Snapshot
}

func (p persisted) chainIs(id string, want facade.ChainStatus) persisted {
	p.t.Helper()
	for _, c := range p.s.Chains {
		if c.ID == id {
			if c.Status != want {
				p.t.Fatalf("chain %s is %s, want %s", id, c.Status, want)
			}
			return p
		}
	}
	p.t.Fatalf("chain %s not on disk", id)
	return p
}

func (p persisted) planIs(want soil.PlanStatus) persisted {
	p.t.Helper()
	if len(p.s.Plans) != 1 || p.s.Plans[0].Status != want {
		p.t.Fatalf("plans on disk: %+v, want one %s", p.s.Plans, want)
	}
	return p
}

func (p persisted) runIs(want climate.RunStatus) persisted {
	p.t.Helper()
	if len(p.s.Runs) != 1 || p.s.Runs[0].Status != want {
		p.t.Fatalf("runs on disk: %+v, want one %s", p.s.Runs, want)
	}
	return p
}

func (p persisted) doseIs(want dosing.DoseStatus) persisted {
	p.t.Helper()
	if len(p.s.Doses) != 1 || p.s.Doses[0].Status != want {
		p.t.Fatalf("doses on disk: %+v, want one %s", p.s.Doses, want)
	}
	return p
}

func (p persisted) tankHolds(ml float64) persisted {
	p.t.Helper()
	if p.s.TankMl != ml {
		p.t.Fatalf("tank = %v ml, want %v", p.s.TankMl, ml)
	}
	return p
}

func (p persisted) hasNothingCommitted() persisted {
	p.t.Helper()
	if len(p.s.Chains)+len(p.s.Plans)+len(p.s.Runs)+len(p.s.Doses) != 0 {
		p.t.Fatalf("expected an empty disk, got %+v", p.s)
	}
	return p
}

func (p persisted) hasNoDoses() persisted {
	p.t.Helper()
	if len(p.s.Doses) != 0 {
		p.t.Fatalf("expected no doses, got %+v", p.s.Doses)
	}
	return p
}

func TestHappyPath_CompletesAndSurvivesRestart(t *testing.T) {
	g := aGreenhouse(t)
	o := g.waters("tomatoes").succeeds()

	g.afterRestart().
		chainIs(o.chain.ID, facade.Completed).
		planIs(soil.PlanCompleted).
		runIs(climate.RunDone).
		doseIs(dosing.DoseDispensed).
		tankHolds(650)
}

func TestBFailsAfterA_OnlyAIsCompensated(t *testing.T) {
	cases := map[string]struct {
		setup func(*greenhouse) *greenhouse
		cause error
	}{
		"low water pressure": {func(g *greenhouse) *greenhouse { return g.withLowWaterPressure() }, climate.ErrLowPressure},
		"stuck valve":        {func(g *greenhouse) *greenhouse { return g.withStuckValve("tomatoes") }, climate.ErrValveStuck},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			g := c.setup(aGreenhouse(t))
			o := g.waters("tomatoes").failsAtStep("B").because(c.cause).compensated("A")

			g.afterRestart().
				chainIs(o.chain.ID, facade.Compensated).
				planIs(soil.PlanReleased).
				runIs(climate.RunAborted).
				hasNoDoses().
				tankHolds(1000)
		})
	}
}

func TestCFailsAfterAandB_BThenAAreCompensatedInOrder(t *testing.T) {
	g := aGreenhouse(t).withNutrientTank(100)
	o := g.waters("tomatoes").failsAtStep("C").because(dosing.ErrTankEmpty).compensated("B", "A")

	if g.drv.IsOpen("tomatoes") {
		t.Fatal("the valve must be closed again after compensating B")
	}
	g.afterRestart().
		chainIs(o.chain.ID, facade.Compensated).
		planIs(soil.PlanReleased).
		runIs(climate.RunAborted).
		doseIs(dosing.DoseFailed).
		tankHolds(100)
}

func TestFirstStepFailures_NothingIsCommitted(t *testing.T) {
	cases := []struct {
		name  string
		setup func(*greenhouse) *greenhouse
		zone  string
		want  error
	}{
		{"soil is already wet", func(g *greenhouse) *greenhouse { return g.withWetSoil("tomatoes") }, "tomatoes", soil.ErrWateringNotNeeded},
		{"unknown zone", func(g *greenhouse) *greenhouse { return g }, "cactus", soil.ErrUnknownZone},
		{"blank zone", func(g *greenhouse) *greenhouse { return g }, "  ", soil.ErrInvalidInput},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := c.setup(aGreenhouse(t))
			g.waters(c.zone).isRejectedBecause(c.want)
			g.afterRestart().hasNothingCommitted()
		})
	}
}

func TestRestartMidChain_RecoverCompensatesWhatIsOnDisk(t *testing.T) {
	g := aGreenhouse(t).crashingAfterB()
	o := g.waters("tomatoes").diesWith(errCrash)

	g.afterRestart().
		chainIs(o.chain.ID, facade.PendingC).
		planIs(soil.PlanReserved).
		runIs(climate.RunRunning).
		hasNoDoses()

	recovered, err := g.admin.Recover()
	g.must(err)
	if len(recovered) != 1 || recovered[0].Status != facade.Compensated ||
		!reflect.DeepEqual(recovered[0].Compensated, []string{"B", "A"}) {
		t.Fatalf("unexpected recovery result: %+v", recovered)
	}
	g.afterRestart().
		chainIs(o.chain.ID, facade.Compensated).
		planIs(soil.PlanReleased).
		runIs(climate.RunAborted)

	again, err := g.admin.Recover()
	g.must(err)
	if len(again) != 0 {
		t.Fatalf("second recovery should do nothing, got %+v", again)
	}
}

func TestManualCompensation(t *testing.T) {
	g := aGreenhouse(t)
	done := g.waters("tomatoes").succeeds()

	if _, err := g.admin.Compensate(done.chain.ID); !errors.Is(err, facade.ErrNotCompensable) {
		t.Fatalf("completed chain: %v", err)
	}
	if _, err := g.admin.Compensate("chain-999"); !errors.Is(err, facade.ErrChainNotFound) {
		t.Fatalf("unknown chain: %v", err)
	}

	g.withDrySoil("tomatoes").crashingAfterB()
	stuck := g.waters("tomatoes").diesWith(errCrash)
	ch, err := g.admin.Compensate(stuck.chain.ID)
	g.must(err)
	if ch.Status != facade.Compensated || !reflect.DeepEqual(ch.Compensated, []string{"B", "A"}) {
		t.Fatalf("manual compensation result: %+v", ch)
	}
	if _, err := g.admin.Compensate(stuck.chain.ID); !errors.Is(err, facade.ErrNotCompensable) {
		t.Fatalf("compensating twice: %v", err)
	}
}
