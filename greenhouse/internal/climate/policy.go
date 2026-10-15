package climate

type VentPolicy struct {
	Min, Max int
	PerLiter float64
}

func DefaultVentPolicy() VentPolicy { return VentPolicy{Min: 20, Max: 100, PerLiter: 1.5} }

func (p VentPolicy) Level(liters float64) int {
	l := int(liters * p.PerLiter)
	if l < p.Min {
		return p.Min
	}
	if l > p.Max {
		return p.Max
	}
	return l
}
