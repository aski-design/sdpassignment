package dosing

import "errors"

var (
	ErrInvalidInput      = errors.New("dosing: invalid input")
	ErrTankEmpty         = errors.New("dosing: not enough nutrient concentrate in the tank")
	ErrIllegalTransition = errors.New("dosing: illegal dose status transition")
)
