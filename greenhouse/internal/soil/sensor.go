package soil

import (
	"fmt"
	"strings"
)

type SensorArray struct{ repo *Repository }

func NewSensorArray(repo *Repository) *SensorArray { return &SensorArray{repo: repo} }

func (s *SensorArray) Record(zone string, moisture float64) error {
	if strings.TrimSpace(zone) == "" || moisture < 0 || moisture > 100 {
		return fmt.Errorf("%w: zone %q, moisture %.1f", ErrInvalidInput, zone, moisture)
	}
	return s.repo.SetReading(zone, moisture)
}

func (s *SensorArray) Moisture(zone string) (float64, error) {
	v, ok := s.repo.Reading(zone)
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownZone, zone)
	}
	return v, nil
}

func (s *SensorArray) All() map[string]float64 { return s.repo.Readings() }
