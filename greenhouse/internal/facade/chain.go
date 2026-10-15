package facade

import (
	"fmt"
	"sync"

	"greenhouse/internal/jsonfile"
)

type ChainStatus string

const (
	PendingB     ChainStatus = "PENDING_B"
	PendingC     ChainStatus = "PENDING_C"
	Completed    ChainStatus = "COMPLETED"
	Compensating ChainStatus = "COMPENSATING"
	Compensated  ChainStatus = "COMPENSATED"
)

type Chain struct {
	ID          string      `json:"id"`
	Zone        string      `json:"zone"`
	Liters      float64     `json:"liters"`
	Status      ChainStatus `json:"status"`
	Error       string      `json:"error,omitempty"`
	Compensated []string    `json:"compensated,omitempty"`
}

func (c Chain) Unfinished() bool {
	switch c.Status {
	case PendingB, PendingC, Compensating:
		return true
	}
	return false
}

type ChainRepo struct {
	mu   sync.Mutex
	path string
	data chainData
}

type chainData struct {
	Seq    int     `json:"seq"`
	Chains []Chain `json:"chains"`
}

func OpenChainRepo(path string) (*ChainRepo, error) {
	r := &ChainRepo{path: path}
	if err := jsonfile.Load(path, &r.data); err != nil {
		return nil, fmt.Errorf("facade: load %s: %w", path, err)
	}
	return r, nil
}

func (r *ChainRepo) save() error { return jsonfile.Save(r.path, r.data) }

func (r *ChainRepo) NewID() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data.Seq++
	return fmt.Sprintf("chain-%d", r.data.Seq), r.save()
}

func (r *ChainRepo) Save(c *Chain) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *c
	cp.Compensated = append([]string(nil), c.Compensated...)
	for i := range r.data.Chains {
		if r.data.Chains[i].ID == c.ID {
			r.data.Chains[i] = cp
			return r.save()
		}
	}
	r.data.Chains = append(r.data.Chains, cp)
	return r.save()
}

func (r *ChainRepo) Get(id string) (*Chain, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.data.Chains {
		if c.ID == id {
			cp := c
			return &cp, true
		}
	}
	return nil, false
}

func (r *ChainRepo) All() []Chain {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Chain(nil), r.data.Chains...)
}
