package dosing

import "fmt"

type Tank struct {
	LevelMl float64 `json:"level_ml"`
}

func (t *Tank) Draw(ml float64) error {
	if ml > t.LevelMl {
		return fmt.Errorf("%w: need %.0f ml, have %.0f ml", ErrTankEmpty, ml, t.LevelMl)
	}
	t.LevelMl -= ml
	return nil
}
