package report

import (
	"fmt"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/sim"
	"html"
	"strings"
)

// GridSVG is deliberately a small writer, not a charting dependency. Each column
// shares one scale across scenarios; numeric labels preserve absolute counts.
func GridSVG(rows []sim.Result, dark bool) string {
	bg, fg, line := "#faf9f6", "#242424", "#dedbd4"
	colors := []string{"#777777", "#343434", "#166294"}
	if dark {
		bg, fg, line = "#181a1c", "#f5f1e8", "#464646"
		colors = []string{"#adadad", "#e4e4e4", "#83c9f4"}
	}
	value := func(r sim.Result, c int) int {
		return []int{r.Lost, r.DuplicateEffects, r.DoubleBilled, r.DrainAbandoned}[c]
	}
	maxima := [4]int{1, 1, 1, 1}
	for _, r := range rows {
		for c := 0; c < 4; c++ {
			if n := value(r, c); n > maxima[c] {
				maxima[c] = n
			}
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="900" height="420" viewBox="0 0 900 420" role="img" aria-labelledby="title desc"><title id="title">Dispatch failures by scenario and strategy</title><desc id="desc">Counts per %d runs. Each column shares a scale. Every bar has a text strategy label and numeric count. Exactly-once has zero invariant violations in all scenarios.</desc><rect width="900" height="420" fill="%s"/><g fill="%s" font-family="system-ui,sans-serif">`, rows[0].Runs, bg, fg)
	fmt.Fprintf(&b, `<text x="14" y="23" font-size="16" font-weight="600">Dispatch failure semantics · %d runs/cell · seed %d</text>`, rows[0].Runs, rows[0].Seed)
	b.WriteString(`<text x="14" y="43" font-size="12">Shared scale per column. retry = at-least-once; exact = exactly-once. A tick and “0” mean zero.</text>`)
	headers := []string{"Lost runs", "Extra tool effects", "Double-billed runs", "Drain-abandoned steps"}
	for c, h := range headers {
		fmt.Fprintf(&b, `<text x="%d" y="68" font-size="12" font-weight="600">%s</text>`, 112+c*194, html.EscapeString(h))
	}
	labels := []string{"naive", "retry", "exact"}
	for i, s := range sim.Scenarios() {
		y := 88 + i*45
		fmt.Fprintf(&b, `<text x="14" y="%d" font-size="13">%s</text><path d="M 12 %d H 888" stroke="%s"/>`, y+18, html.EscapeString(s.Name), y+41, line)
		for c := 0; c < 4; c++ {
			for j, name := range sim.Strategies {
				count := 0
				for _, r := range rows {
					if r.Scenario == s.Name && r.Strategy == name {
						count = value(r, c)
					}
				}
				x := 112 + c*194
				yy := y + j*14
				fmt.Fprintf(&b, `<text x="%d" y="%d" font-size="14">%s</text>`, x, yy+8, labels[j])
				width := float64(count) * 96 / float64(maxima[c])
				barX := x + 39
				if count == 0 {
					fmt.Fprintf(&b, `<path d="M %d %d v 6" stroke="%s" stroke-width="2"/>`, barX, yy+1, colors[j])
				} else {
					fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%.2f" height="5" fill="%s"/>`, barX, yy+2, width, colors[j])
				}
				fmt.Fprintf(&b, `<text x="%d" y="%d" font-size="14">%d</text>`, x+143, yy+8, count)
			}
		}
	}
	b.WriteString(`<text x="14" y="416" font-size="11">Synthetic, virtual-clock experiment. Effects are not provider computations. Raw data: report/results.json.</text></g></svg>`)
	return b.String()
}
