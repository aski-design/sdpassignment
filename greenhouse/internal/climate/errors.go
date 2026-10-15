package climate

import "errors"

var (
	ErrInvalidInput      = errors.New("climate: invalid input")
	ErrLowPressure       = errors.New("climate: water pressure too low")
	ErrValveStuck        = errors.New("climate: valve is stuck")
	ErrIllegalTransition = errors.New("climate: illegal run status transition")
	ErrRunNotFound       = errors.New("climate: run not found")
)
