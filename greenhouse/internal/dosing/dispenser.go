package dosing

import (
	"fmt"
	"strings"
)

type Dispenser struct {
	store  *Store
	policy DosePolicy
}

func NewDispenser(store *Store, policy DosePolicy) *Dispenser {
	return &Dispenser{store: store, policy: policy}
}

func (d *Dispenser) Dispense(ref, zone string, liters float64) (*Dose, error) {
	if strings.TrimSpace(ref) == "" || strings.TrimSpace(zone) == "" || liters <= 0 {
		return nil, fmt.Errorf("%w: ref, zone and positive liters are required", ErrInvalidInput)
	}
	dose := &Dose{Ref: ref, Zone: zone, Ml: d.policy.Amount(liters), Status: DosePending}
	if err := d.store.SaveDose(dose); err != nil {
		return nil, err
	}
	tank := d.store.Tank()
	if err := tank.Draw(dose.Ml); err != nil {
		_ = dose.Fail(err.Error())
		if saveErr := d.store.SaveDose(dose); saveErr != nil {
			return dose, saveErr
		}
		return dose, err
	}
	if err := d.store.SaveTank(tank); err != nil {
		return dose, err
	}
	_ = dose.Dispense()
	return dose, d.store.SaveDose(dose)
}

func (d *Dispenser) SetTankLevel(ml float64) error {
	if ml < 0 {
		return fmt.Errorf("%w: tank level cannot be negative", ErrInvalidInput)
	}
	return d.store.SaveTank(Tank{LevelMl: ml})
}

func (d *Dispenser) TankLevel() float64                 { return d.store.Tank().LevelMl }
func (d *Dispenser) FindByRef(ref string) (*Dose, bool) { return d.store.FindDoseByRef(ref) }
func (d *Dispenser) Doses() []Dose                      { return d.store.Doses() }
