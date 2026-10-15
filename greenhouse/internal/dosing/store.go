package dosing

import (
	"fmt"
	"sync"

	"greenhouse/internal/jsonfile"
)

type Store struct {
	mu   sync.Mutex
	path string
	data storeData
}

type storeData struct {
	Seq   int    `json:"seq"`
	Tank  Tank   `json:"tank"`
	Doses []Dose `json:"doses"`
}

func OpenStore(path string, initialTankMl float64) (*Store, error) {
	s := &Store{path: path, data: storeData{Tank: Tank{LevelMl: initialTankMl}}}
	if err := jsonfile.Load(path, &s.data); err != nil {
		return nil, fmt.Errorf("dosing: load %s: %w", path, err)
	}
	return s, nil
}

func (s *Store) save() error { return jsonfile.Save(s.path, s.data) }

func (s *Store) Tank() Tank {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data.Tank
}

func (s *Store) SaveTank(t Tank) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Tank = t
	return s.save()
}

func (s *Store) SaveDose(d *Dose) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d.ID == "" {
		s.data.Seq++
		d.ID = fmt.Sprintf("dose-%d", s.data.Seq)
		s.data.Doses = append(s.data.Doses, *d)
		return s.save()
	}
	for i := range s.data.Doses {
		if s.data.Doses[i].ID == d.ID {
			s.data.Doses[i] = *d
			return s.save()
		}
	}
	return fmt.Errorf("dosing: dose %s not found", d.ID)
}

func (s *Store) FindDoseByRef(ref string) (*Dose, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range s.data.Doses {
		if d.Ref == ref {
			cp := d
			return &cp, true
		}
	}
	return nil, false
}

func (s *Store) Doses() []Dose {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Dose(nil), s.data.Doses...)
}
