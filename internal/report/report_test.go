package report

import (
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kevinmamaqi/agent-dispatch-lab/internal/sim"
)

func TestFiguresAreWellFormedInBothThemes(t *testing.T) {
	rows := []sim.Result{{Scenario: "storm", Strategy: "exactly-once", Runs: 1000, Seed: 42}}
	for _, dark := range []bool{false, true} {
		for name, s := range map[string]string{"storm": StormSVG(rows, dark), "grid": GridSVG(rows, dark), "mechanism": MechanismSVG(dark)} {
			t.Run(name+map[bool]string{false: "/light", true: "/dark"}[dark], func(t *testing.T) {
				d := xml.NewDecoder(strings.NewReader(s))
				for {
					_, err := d.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
				}
				if !strings.Contains(s, `<desc id="desc">`) {
					t.Fatal("missing accessible description")
				}
			})
		}
	}
}

func TestStormShowsOnlyStormCountsAndExplicitZeros(t *testing.T) {
	rows := []sim.Result{
		{Scenario: "baseline", Strategy: "naive", Lost: 9999},
		{Scenario: "storm", Strategy: "naive", Runs: 10, Completed: 3, Lost: 7, DuplicateEffects: 2, DoubleBilled: 1},
		{Scenario: "storm", Strategy: "at-least-once", Runs: 10, Completed: 10, DuplicateEffects: 4, DoubleBilled: 2},
		{Scenario: "storm", Strategy: "exactly-once", Runs: 10, Completed: 10},
	}
	s := StormSVG(rows, false)
	if strings.Contains(s, "9999") || strings.Count(s, `<g aria-label="`) != 12 {
		t.Fatal("hero must contain three strategies and nine storm measurements")
	}
	for _, label := range []string{`aria-label="Lost runs: 7"`, `aria-label="Extra effects: 4"`, `aria-label="Double-billed: 2"`, `aria-label="exactly-once: 100.0% completed"`} {
		if !strings.Contains(s, label) {
			t.Fatalf("missing measured label %s", label)
		}
	}
	if strings.Count(s, `v 16`) != 4 || strings.Count(s, `>0</text>`) != 5 {
		t.Fatal("four zero measurements plus the zero axis label must stay visible")
	}
}

func TestBatchTimeDoesNotMasqueradeAsRunLatency(t *testing.T) {
	rows := []sim.Result{
		{Scenario: "storm", Strategy: "at-least-once", Runs: 10, BatchMS: 8000, P50MS: 777000},
		{Scenario: "storm", Strategy: "exactly-once", Runs: 10, BatchMS: 4000, P50MS: 999000},
	}
	var terminal bytes.Buffer
	Terminal(&terminal, rows)
	for _, s := range []string{terminal.String(), Markdown(rows)} {
		if strings.Contains(s, "p50") || strings.Contains(s, "p99") {
			t.Fatal("queue-inclusive quantiles returned to the headline table")
		}
	}
	cost := Narratives(rows)["cost"]
	if !strings.Contains(cost, "4.00s") || !strings.Contains(cost, "8.00s") || !strings.Contains(cost, "batch makespan") || strings.Contains(cost, "999.00") {
		t.Fatal("cost must report actual batch makespan, not a relabelled percentile")
	}
}

func TestReportRegeneratesAllFiguresDeterministically(t *testing.T) {
	rows, err := sim.Batch(sim.Config{Runs: 10, Seed: 42, Workers: 4, Scenario: "all", Strategy: "all"})
	if err != nil {
		t.Fatal(err)
	}
	a, b := t.TempDir(), t.TempDir()
	for _, dir := range []string{a, b} {
		if err := Write(dir, rows); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"results.json", "results.md", "results.svg", "results-dark.svg", "results-grid.svg", "results-grid-dark.svg", "mechanism.svg", "mechanism-dark.svg"} {
		x, err := os.ReadFile(filepath.Join(a, name))
		if err != nil {
			t.Fatal(err)
		}
		y, err := os.ReadFile(filepath.Join(b, name))
		if err != nil || !bytes.Equal(x, y) {
			t.Fatalf("%s did not reproduce byte-for-byte: %v", name, err)
		}
	}
}
