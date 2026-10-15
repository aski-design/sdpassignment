package facade

import (
	"path/filepath"

	"greenhouse/internal/climate"
	"greenhouse/internal/dosing"
	"greenhouse/internal/soil"
)

func Assemble(dir string, drv climate.Driver) (*GrowthFacade, *AdminFacade, error) {
	soilRepo, err := soil.OpenRepository(filepath.Join(dir, "soil.json"))
	if err != nil {
		return nil, nil, err
	}
	sensors := soil.NewSensorArray(soilRepo)
	sched := soil.NewScheduler(soilRepo, sensors, soil.DefaultRule())

	climRepo, err := climate.OpenRepository(filepath.Join(dir, "climate.json"))
	if err != nil {
		return nil, nil, err
	}
	ctrl := climate.NewController(climRepo, drv, climate.DefaultVentPolicy(), 1.5)

	store, err := dosing.OpenStore(filepath.Join(dir, "dosing.json"), 1000)
	if err != nil {
		return nil, nil, err
	}
	disp := dosing.NewDispenser(store, dosing.DefaultDosePolicy())

	chains, err := OpenChainRepo(filepath.Join(dir, "chains.json"))
	if err != nil {
		return nil, nil, err
	}

	c := &core{sensors: sensors, sched: sched, ctrl: ctrl, disp: disp, chains: chains}
	return &GrowthFacade{core: c}, &AdminFacade{core: c}, nil
}
