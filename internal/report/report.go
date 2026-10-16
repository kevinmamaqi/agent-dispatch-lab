// Package report writes the measured evidence; it never supplies example data.
package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/sim"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

func percent(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return 100 * float64(n) / float64(d)
}
func Terminal(out io.Writer, rows []sim.Result) {
	previous := ""
	var w *tabwriter.Writer
	for _, r := range rows {
		if previous != r.Scenario {
			if w != nil {
				w.Flush()
			}
			description := ""
			for _, s := range sim.Scenarios() {
				if s.Name == r.Scenario {
					description = s.Description
				}
			}
			fmt.Fprintf(out, "\nscenario: %s (%s) runs=%d seed=%d workers=%d\n", r.Scenario, description, r.Runs, r.Seed, r.Workers)
			w = tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "strategy\tcompleted\tlost\tdup-effects\tdouble-billed\twasted-calls\tabandoned")
			previous = r.Scenario
		}
		fmt.Fprintf(w, "%s\t%.1f%%\t%d\t%d\t%d\t%d\t%d\n", r.Strategy, percent(r.Completed, r.Runs), r.Lost, r.DuplicateEffects, r.DoubleBilled, r.WastedCalls, r.DrainAbandoned)
	}
	if w != nil {
		w.Flush()
	}
}

func Markdown(rows []sim.Result) string {
	var b strings.Builder
	r := rows[0]
	fmt.Fprintf(&b, "Runs per cell: %d · seed: %d · workers: %d. Counts describe effects and completion; see the README’s cost discussion for batch timing.\n\n", r.Runs, r.Seed, r.Workers)
	b.WriteString("| Scenario | Strategy | Completed | Lost | Duplicate effects | Double-billed | Wasted / provider calls | Drain abandoned |\n|---|---|---:|---:|---:|---:|---:|---:|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %.1f%% | %d | %d | %d | %d / %d | %d |\n", r.Scenario, r.Strategy, percent(r.Completed, r.Runs), r.Lost, r.DuplicateEffects, r.DoubleBilled, r.WastedCalls, r.ProviderCalls, r.DrainAbandoned)
	}
	return b.String()
}

func Narratives(rows []sim.Result) map[string]string {
	cell := func(s, n string) sim.Result {
		for _, r := range rows {
			if r.Scenario == s && r.Strategy == n {
				return r
			}
		}
		return sim.Result{}
	}
	retry := cell("storm", "at-least-once")
	exact := cell("storm", "exactly-once")
	dup := cell("dup-5", "at-least-once")
	errN := cell("error-10", "naive")
	errA := cell("error-10", "at-least-once")
	drain := cell("drain", "naive")
	return map[string]string{
		"claim":        fmt.Sprintf("Under combined crashes, duplicate deliveries, hangs, errors and a drain, at-least-once dispatch double-bills %.1f%% of agent runs; idempotent exactly-once dispatch double-bills none, while %.1f%% of its provider calls yield unused results.", percent(retry.DoubleBilled, retry.Runs), percent(exact.WastedCalls, exact.ProviderCalls)),
		"observations": fmt.Sprintf("In `dup-5`, retries show %.1f%% completion but double-bill %d of %d runs. Completion alone misses the failure.\n\nIn `error-10`, retries raise completion from %.1f%% to %.1f%%. Errors before an effect are the case retry handles well.\n\nThe naive drain abandons %d steps and loses %d runs; the idempotent strategy abandons none. These are counts of steps and runs, respectively.\n\nIn `storm`, retry writes %d extra tool effects; exactly-once writes zero extra effects. The simulator counts extra writes, not merely affected keys.", percent(dup.Completed, dup.Runs), dup.DoubleBilled, dup.Runs, percent(errN.Completed, errN.Runs), percent(errA.Completed, errA.Runs), drain.DrainAbandoned, drain.Lost, retry.DuplicateEffects),
		"cost":         fmt.Sprintf("In `storm`, exactly-once makes %d provider calls; %d (%.1f%%) have unused results. Retry makes %d calls with %d unused results. This waste includes failed and interrupted calls, not only an idempotency surcharge. Cached recovery can avoid a second call; an unavailable result can still require new computation.\n\nThe whole %d-run storm batch takes %.2fs of virtual time with exactly-once, versus %.2fs with retry, using %d workers. This is batch makespan through the last scheduled event, including late provider completions; it is not individual run latency or wall time. `results.json` retains `batch_latency_ms` and the queue-inclusive `p50_run_latency_ms` / `p99_run_latency_ms` for billed runs. Those percentiles measure time since the entire batch arrived, not provider response time.", exact.ProviderCalls, exact.WastedCalls, percent(exact.WastedCalls, exact.ProviderCalls), retry.ProviderCalls, retry.WastedCalls, exact.Runs, float64(exact.BatchMS)/1000, float64(retry.BatchMS)/1000, exact.Workers),
		"results":      Markdown(rows),
	}
}

func Write(dir string, rows []sim.Result) error {
	for _, r := range rows {
		if r.Strategy == "exactly-once" && (r.Completed != r.Runs || r.Lost+r.DuplicateEffects+r.DoubleBilled+r.DrainAbandoned != 0) {
			return fmt.Errorf("exactly-once invariant failed in %s; refusing a success claim", r.Scenario)
		}
	}
	if len(rows) != 21 {
		return fmt.Errorf("reports require all seven scenarios and three strategies")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	files := []struct {
		name string
		data []byte
	}{
		{"results.json", data},
		{"results.md", []byte(Markdown(rows))},
		{"results.svg", []byte(StormSVG(rows, false))},
		{"results-dark.svg", []byte(StormSVG(rows, true))},
		{"results-grid.svg", []byte(GridSVG(rows, false))},
		{"results-grid-dark.svg", []byte(GridSVG(rows, true))},
		{"mechanism.svg", []byte(MechanismSVG(false))},
		{"mechanism-dark.svg", []byte(MechanismSVG(true))},
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(dir, file.name), file.data, 0644); err != nil {
			return err
		}
	}
	// README is written last during implementation. Thereafter the same command
	// refreshes its measured blocks, so claims cannot drift from the raw data.
	readme, err := os.ReadFile("README.md")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, name := range []string{"claim", "results", "observations", "cost"} {
		start := []byte("<!-- " + name + ":start -->")
		end := []byte("<!-- " + name + ":end -->")
		a := bytes.Index(readme, start)
		b := bytes.Index(readme, end)
		if a < 0 || b < a {
			return fmt.Errorf("README missing measured block %q", name)
		}
		next := append([]byte{}, readme[:a+len(start)]...)
		next = append(next, []byte("\n"+Narratives(rows)[name]+"\n")...)
		next = append(next, readme[b:]...)
		readme = next
	}
	return os.WriteFile("README.md", readme, 0644)
}
