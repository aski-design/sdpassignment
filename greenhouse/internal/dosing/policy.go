package dosing

type DosePolicy struct{ MlPerLiter float64 }

func DefaultDosePolicy() DosePolicy { return DosePolicy{MlPerLiter: 5} }

func (p DosePolicy) Amount(liters float64) float64 { return liters * p.MlPerLiter }
