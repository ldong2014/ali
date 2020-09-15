package gui

import (
	"context"
	"time"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"

	"github.com/nakabonne/ali/attacker"
)

type drawer struct {
	widgets  *widgets
	chartCh  chan *attacker.Result
	gaugeCh  chan bool
	reportCh chan string

	// aims to avoid to perform multiple `redrawChart`.
	chartDrawing bool
}

// redrawChart appends entities as soon as a result arrives.
// Given maxSize, then it can be pre-allocated.
// TODO: In the future, multiple charts including bytes-in/out etc will be re-drawn.
func (d *drawer) redrawChart(ctx context.Context, maxSize int) {
	values := make([]float64, 0, maxSize)
	d.chartDrawing = true
L:
	for {
		select {
		case <-ctx.Done():
			break L
		case res := <-d.chartCh:
			if res.End {
				d.gaugeCh <- true
				break L
			}
			d.gaugeCh <- false
			values = append(values, float64(res.Latency/time.Millisecond))
			d.widgets.latencyChart.Series("latency", values,
				linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(87))),
				linechart.SeriesXLabels(map[int]string{
					0: "req",
				}),
			)
		}
	}
	d.chartDrawing = false
}

func (d *drawer) redrawGauge(ctx context.Context, maxSize int) {
	var count float64
	size := float64(maxSize)
	d.widgets.progressGauge.Percent(0)
	for {
		select {
		case <-ctx.Done():
			return
		case end := <-d.gaugeCh:
			if end {
				return
			}
			count++
			d.widgets.progressGauge.Percent(int(count / size * 100))
		}
	}
}

func (d *drawer) redrawReport(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case report := <-d.reportCh:
			d.widgets.reportText.Write(report, text.WriteReplace())
		}
	}
}
