package report

import (
	"fmt"
	"html"
	"strings"
)

// MechanismSVG illustrates the after-effect/before-log crash window, not a
// trace sampled from the batch. Frames share a ten-second SMIL timeline. The
// final frame is also the no-SMIL and reduced-motion fallback.
func MechanismSVG(dark bool) string {
	bg, fg, muted, line, accent := "#faf9f6", "#242424", "#65635f", "#dedbd4", "#166294"
	if dark {
		bg, fg, muted, line, accent = "#181a1c", "#f5f1e8", "#b9b5ad", "#464646", "#83c9f4"
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="900" height="560" viewBox="0 0 900 560" role="img" aria-labelledby="title desc"><title id="title">The write succeeded. The acknowledgement did not.</title><desc id="desc">One tool step, three strategies. A worker crashes after writing an effect but before recording completion. Naive loses progress despite the existing effect. Retry redelivers and writes twice. Exactly-once grants a new version, rejects a stale duplicate claim, reuses the cached result, deduplicates the write and commits. This is an illustrative ten-second loop, with a static final frame for reduced motion.</desc><style>.still{display:none}@media(prefers-reduced-motion:reduce){.motion{display:none}.still{display:inline}}</style><rect width="900" height="560" fill="%s"/><g fill="%s" font-family="system-ui,sans-serif">`, bg, fg)
	b.WriteString(`<text x="24" y="34" font-size="24" font-weight="650">The write succeeded. The acknowledgement didn’t.</text>`)
	fmt.Fprintf(&b, `<text x="24" y="61" font-size="15" fill="%s">One tool step · crash after effect, before completion is logged · same key on redelivery</text>`, muted)
	for i, label := range []string{"Queue", "Worker", "Provider", "Ledger"} {
		fmt.Fprintf(&b, `<text x="%d" y="99" text-anchor="middle" font-size="16" font-weight="600">%s</text>`, []int{271, 433, 593, 785}[i], label)
	}
	for i, label := range []string{"naive", "at-least-once", "exactly-once"} {
		y, color := 144+i*132, fg
		if i == 2 {
			color = accent
		}
		fmt.Fprintf(&b, `<text x="24" y="%d" font-size="19" font-weight="650" fill="%s">%s</text><text x="24" y="%d" font-size="15" fill="%s">%s</text>`, y-3, color, label, y+22, muted, []string{"No recovery", "Retry, no key check", "Key + version + log"}[i])
		for j, x := range []int{228, 380, 540, 716} {
			width := []int{86, 106, 106, 138}[j]
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="48" rx="8" fill="%s" stroke="%s"/>`, x, y-24, width, bg, line)
			if j < 3 {
				next := []int{380, 540, 716}[j]
				fmt.Fprintf(&b, `<path d="M %d %d H %d m -5 -4 l 5 4 -5 4" fill="none" stroke="%s"/>`, x+width, y, next-6, line)
			}
		}
		if i < 2 {
			fmt.Fprintf(&b, `<path d="M 24 %d H 876" stroke="%s"/>`, y+78, line)
		}
	}
	// All text in a frame changes together: one cause and one observable outcome.
	frame := func(n int) string {
		var f strings.Builder
		states := [3][4]string{}
		captions := [3]string{}
		for i := 0; i < 3; i++ {
			switch n {
			case 0:
				states[i] = [4]string{"step", "running", "call 1", "0 writes"}
				captions[i] = "Dispatch the tool step; wait for the result."
			case 1:
				states[i] = [4]string{"no ack", "crashed", "cached", "1 write"}
				captions[i] = "Effect exists; completion has not been recorded."
				if i == 2 {
					captions[i] = "Effect exists; the step log still says provider_called."
				}
			default:
				states[i] = [4]string{"empty", "dead", "1 call", "1 write"}
				captions[i] = "No redelivery: the run loses progress; the first row remains."
			}
		}
		if n == 2 {
			states[1] = [4]string{"retry", "running", "call 2", "1 write"}
			states[2] = [4]string{"v1 + v0", "v1 wins", "cache hit", "1 write"}
			captions[1] = "Redelivery repeats the provider call and the unkeyed write."
			captions[2] = "v0 is fenced; the v1 owner recovers the cached result."
		}
		if n == 3 {
			states[1] = [4]string{"acked", "done", "2 calls", "2 writes"}
			states[2] = [4]string{"acked", "committed", "1 call", "1 write"}
			captions[1] = "Progress restored. The same tool effect happened twice."
			captions[2] = "Keyed write deduplicated → commit. Billing waits for all steps."
		}
		for i := 0; i < 3; i++ {
			y, color := 144+i*132, fg
			if i == 2 {
				color = accent
			}
			for j, text := range states[i] {
				fmt.Fprintf(&f, `<text x="%d" y="%d" text-anchor="middle" font-size="16" font-weight="600" fill="%s">%s</text>`, []int{271, 433, 593, 785}[j], y+6, color, html.EscapeString(text))
			}
			fmt.Fprintf(&f, `<text x="228" y="%d" font-size="15" fill="%s">%s</text>`, y+54, color, html.EscapeString(captions[i]))
		}
		fmt.Fprintf(&f, `<text x="24" y="519" font-size="18" font-weight="600">%s</text>`, []string{"1 / 4  Dispatch one step", "2 / 4  Crash in the ambiguous window", "3 / 4  Redeliver, then replay or recover", "4 / 4  Lost progress / duplicate effect / one committed effect"}[n])
		return f.String()
	}
	b.WriteString(`<g class="motion">`)
	for n := 0; n < 4; n++ {
		opacity := 0
		if n == 3 {
			opacity = 1
		}
		fmt.Fprintf(&b, `<g opacity="%d"><animate attributeName="opacity" dur="10s" repeatCount="indefinite" calcMode="discrete" keyTimes="0;0.23;0.43;0.72;1" values="%s"/>%s`, opacity, []string{"1;0;0;0;0", "0;1;0;0;0", "0;0;1;0;0", "0;0;0;1;1"}[n], frame(n))
		if n == 0 || n == 2 {
			for i := 0; i < 3; i++ {
				if n == 2 && i == 0 {
					continue // naive does not manufacture a recovery delivery
				}
				y, color := 144+i*132, fg
				if i == 2 {
					color = accent
				}
				times := "0;0.22;0.23;1"
				if n == 2 {
					times = "0;0.43;0.68;1"
				}
				points := "0;1;1;1"
				if n == 2 {
					points = "0;0;1;1"
				}
				fmt.Fprintf(&b, `<circle r="5" fill="%s"><animateMotion dur="10s" repeatCount="indefinite" calcMode="linear" keyTimes="%s" keyPoints="%s" path="M 271 %d H 785"/></circle>`, color, times, points, y)
				if n == 2 && i == 2 {
					fmt.Fprintf(&b, `<circle r="5" fill="%s" stroke="%s" stroke-width="2"><animateMotion dur="10s" repeatCount="indefinite" calcMode="linear" keyTimes="0;0.48;0.56;1" keyPoints="0;0;1;1" path="M 271 %d H 433"/></circle>`, bg, accent, y-32)
				}
			}
		}
		b.WriteString(`</g>`)
	}
	fmt.Fprintf(&b, `</g><g class="still">%s</g><text x="24" y="546" font-size="13" fill="%s">Illustration of one crash window. A cache miss can still require another physical provider call.</text></g></svg>`, frame(3), muted)
	return b.String()
}
