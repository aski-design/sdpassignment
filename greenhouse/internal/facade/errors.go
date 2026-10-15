package facade

import (
	"errors"
	"fmt"
)

var (
	ErrChainNotFound  = errors.New("facade: chain not found")
	ErrNotCompensable = errors.New("facade: chain is already finished and cannot be compensated")
)

type StepError struct {
	Step        string
	Err         error
	Compensated []string
}

func (e *StepError) Error() string {
	return fmt.Sprintf("step %s failed: %v (compensated: %v)", e.Step, e.Err, e.Compensated)
}

func (e *StepError) Unwrap() error { return e.Err }
