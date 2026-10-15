package climate

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
	Seq  int   `json:"seq"`
	Runs []Run `json:"runs"`
}

func OpenRepository(path string) (*Repository, error) {
	r := &Repository{path: path}
	if err := jsonfile.Load(path, &r.data); err != nil {
		return nil, fmt.Errorf("climate: load %s: %w", path, err)
	}
	return r, nil
}

func (r *Repository) save() error { return jsonfile.Save(r.path, r.data) }

func (r *Repository) SaveRun(run *Run) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if run.ID == "" {
		r.data.Seq++
		run.ID = fmt.Sprintf("run-%d", r.data.Seq)
		r.data.Runs = append(r.data.Runs, *run)
		return r.save()
	}
	for i := range r.data.Runs {
		if r.data.Runs[i].ID == run.ID {
			r.data.Runs[i] = *run
			return r.save()
		}
	}
	return ErrRunNotFound
}

func (r *Repository) GetRun(id string) (*Run, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range r.data.Runs {
		if x.ID == id {
			cp := x
			return &cp, true
		}
	}
	return nil, false
}

func (r *Repository) FindRunByRef(ref string) (*Run, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range r.data.Runs {
		if x.Ref == ref {
			cp := x
			return &cp, true
		}
	}
	return nil, false
}

func (r *Repository) Runs() []Run {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Run(nil), r.data.Runs...)
}
