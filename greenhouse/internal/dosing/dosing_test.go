package dosing_test

import (
	"errors"
	"path/filepath"
	"testing"

	"greenhouse/internal/dosing"
)

func newDispenser(t *testing.T, tankMl float64) (*dosing.Dispenser, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dosing.json")
	store, err := dosing.OpenStore(path, tankMl)
	if err != nil {
		t.Fatal(err)
	}
	return dosing.NewDispenser(store, dosing.DefaultDosePolicy()), path
}

func TestDispenseDrawsFromTank(t *testing.T) {
	d, _ := newDispenser(t, 1000)
	dose, err := d.Dispense("r1", "z1", 10)
	if err != nil || dose.Ml != 50 || dose.Status != dosing.DoseDispensed {
		t.Fatalf("dose=%+v err=%v", dose, err)
	}
	if d.TankLevel() != 950 {
		t.Fatalf("tank = %v, want 950", d.TankLevel())
	}
}

func TestTankEmptyFailsTheDoseAndKeepsTheTank(t *testing.T) {
	d, _ := newDispenser(t, 10)
	dose, err := d.Dispense("r1", "z1", 10)
	if !errors.Is(err, dosing.ErrTankEmpty) || dose.Status != dosing.DoseFailed {
		t.Fatalf("dose=%+v err=%v", dose, err)
	}
	if d.TankLevel() != 10 {
		t.Fatalf("tank changed to %v", d.TankLevel())
	}
	if got, ok := d.FindByRef("r1"); !ok || got.Status != dosing.DoseFailed {
		t.Fatalf("failed dose not persisted: %+v", got)
	}
}

func TestInvalidInputAndTransitions(t *testing.T) {
	d, _ := newDispenser(t, 100)
	if _, err := d.Dispense("", "z", 1); !errors.Is(err, dosing.ErrInvalidInput) {
		t.Fatalf("blank ref: %v", err)
	}
	if _, err := d.Dispense("r", "z", 0); !errors.Is(err, dosing.ErrInvalidInput) {
		t.Fatalf("zero liters: %v", err)
	}
	if err := d.SetTankLevel(-1); !errors.Is(err, dosing.ErrInvalidInput) {
		t.Fatalf("negative tank: %v", err)
	}
	dose := dosing.Dose{Status: dosing.DosePending}
	_ = dose.Dispense()
	if err := dose.Fail("x"); !errors.Is(err, dosing.ErrIllegalTransition) {
		t.Fatalf("fail after dispense: %v", err)
	}
	if err := dose.Dispense(); !errors.Is(err, dosing.ErrIllegalTransition) {
		t.Fatalf("dispense twice: %v", err)
	}
}

func TestStateSurvivesReopen(t *testing.T) {
	d, path := newDispenser(t, 1000)
	_, _ = d.Dispense("r1", "z1", 10)

	store2, err := dosing.OpenStore(path, 999)
	if err != nil {
		t.Fatal(err)
	}
	d2 := dosing.NewDispenser(store2, dosing.DefaultDosePolicy())
	if d2.TankLevel() != 950 || len(d2.Doses()) != 1 {
		t.Fatalf("tank=%v doses=%d", d2.TankLevel(), len(d2.Doses()))
	}
}

func TestDosePolicy(t *testing.T) {
	if dosing.DefaultDosePolicy().Amount(10) != 50 {
		t.Fatal("amount is wrong")
	}
}
