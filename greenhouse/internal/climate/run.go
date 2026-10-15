package climate

import "fmt"

type RunStatus string

const (
	RunPending RunStatus = "PENDING"
	RunRunning RunStatus = "RUNNING"
	RunDone    RunStatus = "DONE"
	RunAborted RunStatus = "ABORTED"
)

type Run struct {
	ID        string    `json:"id"`
	Ref       string    `json:"ref"`
	Zone      string    `json:"zone"`
	Liters    float64   `json:"liters"`
	VentLevel int       `json:"vent_level"`
	Status    RunStatus `json:"status"`
	Reason    string    `json:"reason,omitempty"`
}

func (r *Run) Active() bool { return r.Status == RunPending || r.Status == RunRunning }

func (r *Run) Start() error  { return r.move(RunPending, RunRunning) }
func (r *Run) Finish() error { return r.move(RunRunning, RunDone) }

func (r *Run) Abort(reason string) error {
	if !r.Active() {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalTransition, r.Status, RunAborted)
	}
	r.Status = RunAborted
	r.Reason = reason
	return nil
}

func (r *Run) move(from, to RunStatus) error {
	if r.Status != from {
		return fmt.Errorf("%w: %s -> %s", ErrIllegalTransition, r.Status, to)
	}
	r.Status = to
	return nil
}
