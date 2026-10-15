package climate

import (
	"fmt"
	"sync"
)

type Driver interface {
	Pressure() float64
	OpenValve(zone string, liters float64) error
	CloseValve(zone string) error
	SetVent(level int) error
}

type SimDriver struct {
	mu       sync.Mutex
	pressure float64
	stuck    map[string]bool
	open     map[string]bool
	vent     int
}

func NewSimDriver() *SimDriver {
	return &SimDriver{pressure: 3.0, stuck: map[string]bool{}, open: map[string]bool{}}
}

func (d *SimDriver) SetPressure(bar float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.pressure = bar
}

func (d *SimDriver) SetStuck(zone string, stuck bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stuck[zone] = stuck
}

func (d *SimDriver) Pressure() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.pressure
}

func (d *SimDriver) IsOpen(zone string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.open[zone]
}

func (d *SimDriver) VentLevel() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.vent
}

func (d *SimDriver) OpenValve(zone string, _ float64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stuck[zone] {
		return fmt.Errorf("%w: zone %q", ErrValveStuck, zone)
	}
	d.open[zone] = true
	return nil
}

func (d *SimDriver) CloseValve(zone string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.open, zone)
	return nil
}

func (d *SimDriver) SetVent(level int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.vent = level
	return nil
}
