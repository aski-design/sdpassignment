package soil

import "fmt"

type PlanStatus string

const (
	PlanReserved  PlanStatus = "RESERVED"
	PlanCompleted PlanStatus = "COMPLETED"
	PlanReleased  PlanStatus = "RELEASED"
)

type WateringPlan struct {
	ID     string     `json:"id"`
	Ref    string     `json:"ref"`
	Zone   string     `json:"zone"`
	Liters float64    `json:"liters"`
	Status PlanStatus `json:"status"`
}

func (p *WateringPlan) Complete() error { return p.moveFromReserved(PlanCompleted) }
func (p *WateringPlan) Release() error  { return p.moveFromReserved(PlanReleased) }

func (p *WateringPlan) moveFromReserved(to PlanStatus) error {
	if p.Status != PlanReserved {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalTransition, p.Status, to)
	}
	p.Status = to
	return nil
}
