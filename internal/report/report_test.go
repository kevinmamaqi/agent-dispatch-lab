package report

import (
	"encoding/xml"
	"github.com/kevinmamaqi/agent-dispatch-lab/internal/sim"
	"io"
	"strings"
	"testing"
)

func TestSVGIsWellFormedAndLabelsZero(t *testing.T) {
	rows := []sim.Result{{Scenario: "baseline", Strategy: "exactly-once", Runs: 1000, Seed: 42}}
	for _, dark := range []bool{false, true} {
		s := SVG(rows, dark)
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
		if !strings.Contains(s, `v 6`) || !strings.Contains(s, `>0</text>`) {
			t.Fatal("zero became invisible")
		}
	}
}
