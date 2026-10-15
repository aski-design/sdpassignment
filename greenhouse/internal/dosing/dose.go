package dosing

import "fmt"

type DoseStatus string

const (
	DosePending   DoseStatus = "PENDING"
	DoseDispensed DoseStatus = "DISPENSED"
	DoseFailed    DoseStatus = "FAILED"
)

type Dose struct {
	ID     string     `json:"id"`
	Ref    string     `json:"ref"`
	Zone   string     `json:"zone"`
	Ml     float64    `json:"ml"`
	Status DoseStatus `json:"status"`
	Reason string     `json:"reason,omitempty"`
}

func (d *Dose) Dispense() error {
	if d.Status != DosePending {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalTransition, d.Status, DoseDispensed)
	}
	d.Status = DoseDispensed
	return nil
}

func (d *Dose) Fail(reason string) error {
	if d.Status != DosePending {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalTransition, d.Status, DoseFailed)
	}
	d.Status = DoseFailed
	d.Reason = reason
	return nil
}
