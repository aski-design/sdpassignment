package soil

type MoistureRule struct {
	Threshold        float64
	Target           float64
	LitersPerPercent float64
}

func DefaultRule() MoistureRule {
	return MoistureRule{Threshold: 35, Target: 60, LitersPerPercent: 2}
}

func (r MoistureRule) NeedsWater(moisture float64) bool { return moisture < r.Threshold }

func (r MoistureRule) LitersNeeded(moisture float64) float64 {
	if moisture >= r.Target {
		return 0
	}
	return (r.Target - moisture) * r.LitersPerPercent
}
