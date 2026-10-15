package soil

import "errors"

var (
	ErrInvalidInput      = errors.New("soil: invalid input")
	ErrUnknownZone       = errors.New("soil: no sensor reading for zone")
	ErrWateringNotNeeded = errors.New("soil: moisture is sufficient, watering not needed")
	ErrIllegalTransition = errors.New("soil: illegal plan status transition")
	ErrPlanNotFound      = errors.New("soil: plan not found")
)
