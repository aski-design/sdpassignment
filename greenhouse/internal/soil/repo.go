package soil

import (
	"fmt"
	"sync"

	"greenhouse/internal/jsonfile"
)

type Repository struct {
	mu   sync.Mutex
	path string
	data repoData
}

type repoData struct {
	Seq      int                `json:"seq"`
	Readings map[string]float64 `json:"readings"`
	Plans    []WateringPlan     `json:"plans"`
}

func OpenRepository(path string) (*Repository, error) {
	r := &Repository{path: path}
	if err := jsonfile.Load(path, &r.data); err != nil {
		return nil, fmt.Errorf("soil: load %s: %w", path, err)
	}
	if r.data.Readings == nil {
		r.data.Readings = map[string]float64{}
	}
	return r, nil
}

func (r *Repository) save() error { return jsonfile.Save(r.path, r.data) }

func (r *Repository) SetReading(zone string, moisture float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.Readings[zone] = moisture
	return r.save()
}

func (r *Repository) Reading(zone string) (float64, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.data.Readings[zone]
	return v, ok
}

func (r *Repository) Readings() map[string]float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]float64, len(r.data.Readings))
	for k, v := range r.data.Readings {
		out[k] = v
	}
	return out
}

func (r *Repository) SavePlan(p *WateringPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p.ID == "" {
		r.data.Seq++
		p.ID = fmt.Sprintf("plan-%d", r.data.Seq)
		r.data.Plans = append(r.data.Plans, *p)
		return r.save()
	}
	for i := range r.data.Plans {
		if r.data.Plans[i].ID == p.ID {
			r.data.Plans[i] = *p
			return r.save()
		}
	}
	return ErrPlanNotFound
}

func (r *Repository) GetPlan(id string) (*WateringPlan, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.data.Plans {
		if p.ID == id {
			cp := p
			return &cp, true
		}
	}
	return nil, false
}

func (r *Repository) FindPlanByRef(ref string) (*WateringPlan, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.data.Plans {
		if p.Ref == ref {
			cp := p
			return &cp, true
		}
	}
	return nil, false
}

func (r *Repository) Plans() []WateringPlan {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]WateringPlan(nil), r.data.Plans...)
}
