package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"greenhouse/internal/climate"
	"greenhouse/internal/facade"
)

func main() {
	dir := flag.String("data", "data", "directory with the persisted state")
	reset := flag.Bool("reset", false, "delete all persisted state before starting")
	flag.Parse()

	if *reset {
		if err := os.RemoveAll(*dir); err != nil {
			fmt.Println("cannot reset:", err)
			os.Exit(1)
		}
		fmt.Println("persisted state cleared")
	}

	drv := climate.NewSimDriver()
	growth, admin, err := facade.Assemble(*dir, drv)
	if err != nil {
		fmt.Println("startup error:", err)
		os.Exit(1)
	}
	seedIfEmpty(admin)
	warnUnfinished(admin)

	in := bufio.NewScanner(os.Stdin)
	for {
		printMenu()
		switch ask(in, "> ") {
		case "1":
			ch, err := growth.WaterZone(ask(in, "zone: "))
			report(ch, err)
		case "2":
			growth.AfterB = func() error {
				fmt.Println("*** simulated crash right after step B (process killed) ***")
				fmt.Println("Restart the program and look at option 7, then option 8.")
				os.Exit(2)
				return nil
			}
			ch, err := growth.WaterZone(ask(in, "zone: "))
			growth.AfterB = nil
			report(ch, err)
		case "3":
			zone := ask(in, "zone: ")
			if v, ok := number(in, "moisture % (0-100): "); ok {
				printErr(admin.RecordMoisture(zone, v))
			}
		case "4":
			if v, ok := number(in, "water pressure, bar (min is 1.5): "); ok {
				drv.SetPressure(v)
				fmt.Printf("pressure is now %.1f bar\n", v)
			}
		case "5":
			zone := ask(in, "zone: ")
			stuck := strings.EqualFold(ask(in, "valve stuck? (y/n): "), "y")
			drv.SetStuck(zone, stuck)
			fmt.Printf("valve of %q stuck=%v\n", zone, stuck)
		case "6":
			if v, ok := number(in, "tank level, ml: "); ok {
				printErr(admin.SetTankLevel(v))
			}
		case "7":
			printSnapshot(admin.Snapshot())
		case "8":
			done, err := admin.Recover()
			for _, c := range done {
				fmt.Printf("recovered %s -> %s, undone: %v\n", c.ID, c.Status, c.Compensated)
			}
			if len(done) == 0 && err == nil {
				fmt.Println("nothing to recover")
			}
			printErr(err)
		case "9":
			ch, err := admin.Compensate(ask(in, "chain id: "))
			if err == nil {
				fmt.Printf("%s -> %s, undone: %v\n", ch.ID, ch.Status, ch.Compensated)
			}
			printErr(err)
		case "0":
			return
		default:
			fmt.Println("unknown option")
		}
	}
}

func printMenu() {
	fmt.Print(`
--- Smart Greenhouse ---
 1  water a zone (full chain A->B->C)
 2  water a zone and CRASH right after B
 3  set soil moisture of a zone
 4  set water pressure         (fault for step B)
 5  stick / free a zone valve  (fault for step B)
 6  set nutrient tank level    (fault for step C)
 7  show persisted state
 8  recover unfinished chains
 9  compensate one chain manually
 0  exit
`)
}

func ask(in *bufio.Scanner, prompt string) string {
	fmt.Print(prompt)
	if !in.Scan() {
		os.Exit(0)
	}
	return strings.TrimSpace(in.Text())
}

func number(in *bufio.Scanner, prompt string) (float64, bool) {
	v, err := strconv.ParseFloat(ask(in, prompt), 64)
	if err != nil {
		fmt.Println("not a number")
		return 0, false
	}
	return v, true
}

func printErr(err error) {
	if err != nil {
		fmt.Println("error:", err)
	}
}

func report(ch *facade.Chain, err error) {
	var se *facade.StepError
	switch {
	case err == nil:
		fmt.Printf("OK: %s COMPLETED (%.0f L)\n", ch.ID, ch.Liters)
	case errors.As(err, &se):
		fmt.Printf("FAILED at step %s: %v\n  undone (in order): %v -> %s is %s\n", se.Step, se.Err, se.Compensated, ch.ID, ch.Status)
	default:
		fmt.Println("REJECTED:", err)
	}
}

func seedIfEmpty(admin *facade.AdminFacade) {
	if len(admin.Snapshot().Moisture) > 0 {
		return
	}
	for zone, pct := range map[string]float64{"tomatoes": 25, "lettuce": 30, "herbs": 70} {
		printErr(admin.RecordMoisture(zone, pct))
	}
}

func warnUnfinished(admin *facade.AdminFacade) {
	for _, c := range admin.Snapshot().Chains {
		if c.Unfinished() {
			fmt.Printf("WARNING: %s is %s on disk (use option 8)\n", c.ID, c.Status)
		}
	}
}

func printSnapshot(s facade.Snapshot) {
	fmt.Println("== chains")
	for _, c := range s.Chains {
		fmt.Printf("  %-8s %-12s zone=%s %.0fL undone=%v %s\n", c.ID, c.Status, c.Zone, c.Liters, c.Compensated, c.Error)
	}
	fmt.Println("== A: plans")
	for _, p := range s.Plans {
		fmt.Printf("  %-7s ref=%s %-9s %.0fL\n", p.ID, p.Ref, p.Status, p.Liters)
	}
	fmt.Println("== A: soil moisture")
	zones := make([]string, 0, len(s.Moisture))
	for z := range s.Moisture {
		zones = append(zones, z)
	}
	sort.Strings(zones)
	for _, z := range zones {
		fmt.Printf("  %-10s %.0f%%\n", z, s.Moisture[z])
	}
	fmt.Println("== B: runs")
	for _, r := range s.Runs {
		fmt.Printf("  %-6s ref=%s %-8s vent=%d%% %s\n", r.ID, r.Ref, r.Status, r.VentLevel, r.Reason)
	}
	fmt.Println("== C: doses")
	for _, d := range s.Doses {
		fmt.Printf("  %-7s ref=%s %-9s %.0f ml %s\n", d.ID, d.Ref, d.Status, d.Ml, d.Reason)
	}
	fmt.Printf("== C: nutrient tank: %.0f ml\n", s.TankMl)
}
