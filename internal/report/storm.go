package report

import (
	"fmt"
	"html"
	"strings"

	"github.com/kevinmamaqi/agent-dispatch-lab/internal/sim"
)

// StormSVG puts one scenario on one shared count scale. Short bars and zeros
// keep their numeric labels; a zero is a tick, never an invented minimum bar.
func StormSVG(rows []sim.Result, dark bool) string {
	bg, fg, muted, line, ink := "#faf9f6", "#242424", "#65635f", "#dedbd4", "#ffffff"
	colors := []string{"#777777", "#343434", "#166294"}
	if dark {
		bg, fg, muted, line, ink = "#181a1c", "#f5f1e8", "#b9b5ad", "#3d4144", "#181a1c"
		colors = []string{"#adadad", "#e4e4e4", "#83c9f4"}
	}
	storm := make([]sim.Result, len(sim.Strategies))
	var meta sim.Result
	maximum := 1
	for j, name := range sim.Strategies {
		for _, r := range rows {
			if r.Scenario == "storm" && r.Strategy == name {
				meta = r
				storm[j] = r
				for _, n := range []int{r.Lost, r.DuplicateEffects, r.DoubleBilled} {
					if n > maximum {
						maximum = n
					}
				}
			}
		}
	}
	maximum = (maximum + 99) / 100 * 100
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="900" height="420" viewBox="0 0 900 420" role="img" aria-labelledby="title desc"><title id="title">Storm: completion and billing are separate guarantees</title><desc id="desc">%d runs per strategy. Nine bars on one shared count scale show lost runs, extra tool writes and double-billed runs. Each strategy also reports its completion rate. Zero is a baseline tick.</desc><rect width="900" height="420" fill="%s"/><g fill="%s" font-family="system-ui,sans-serif">`, meta.Runs, bg, fg)
	b.WriteString(`<text x="24" y="34" font-size="24" font-weight="650">Retry restores completion. It can repeat billing.</text>`)
	fmt.Fprintf(&b, `<text x="24" y="61" font-size="15" fill="%s">storm · %d runs · %d workers · seed %d · crashes, duplicates, hangs, errors + drain</text>`, muted, meta.Runs, meta.Workers, meta.Seed)
	for i := 0; i <= 3; i++ {
		x := 355 + float64(i)*470/3
		fmt.Fprintf(&b, `<path d="M %.2f 100 V 366" stroke="%s"/><text x="%.2f" y="90" text-anchor="middle" font-size="13" fill="%s">%d</text>`, x, line, x, muted, i*maximum/3)
	}
	labels := []string{"Lost runs", "Extra effects", "Double-billed"}
	for j, r := range storm {
		y := 106 + j*96
		fmt.Fprintf(&b, `<g aria-label="%s: %.1f%% completed"><text x="24" y="%d" font-size="19" font-weight="650" fill="%s">%s</text><text x="24" y="%d" font-size="15" fill="%s">%.1f%% complete</text>`, html.EscapeString(sim.Strategies[j]), percent(r.Completed, r.Runs), y+20, colors[j], html.EscapeString(sim.Strategies[j]), y+44, muted, percent(r.Completed, r.Runs))
		for c, count := range []int{r.Lost, r.DuplicateEffects, r.DoubleBilled} {
			yy := y + c*24
			width := float64(count) * 470 / float64(maximum)
			fmt.Fprintf(&b, `<g aria-label="%s: %d"><text x="337" y="%d" text-anchor="end" font-size="15">%s</text>`, labels[c], count, yy+13, labels[c])
			if count == 0 {
				fmt.Fprintf(&b, `<path d="M 355 %d v 16" stroke="%s" stroke-width="2"/>`, yy, colors[j])
			} else {
				fmt.Fprintf(&b, `<rect x="355" y="%d" width="%.2f" height="17" rx="2" fill="%s"/>`, yy, width, colors[j])
			}
			x, anchor, color := 365+width, "start", fg
			if width >= 36 {
				x, anchor, color = 355+width-7, "end", ink
			}
			if j == 2 && count == 0 {
				color = colors[j]
			}
			fmt.Fprintf(&b, `<text x="%.2f" y="%d" text-anchor="%s" font-size="15" font-weight="650" fill="%s">%d</text></g>`, x, yy+13, anchor, color, count)
		}
		b.WriteString(`</g>`)
		if j < 2 {
			fmt.Fprintf(&b, `<path d="M 24 %d H 876" stroke="%s"/>`, y+81, line)
		}
	}
	fmt.Fprintf(&b, `<text x="24" y="399" font-size="13" fill="%s">Lost / billed: runs. Extra effects: writes beyond the first. Shared scale; counts per %d runs.</text></g></svg>`, muted, meta.Runs)
	return b.String()
}
