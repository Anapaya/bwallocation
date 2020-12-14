// Copyright 2020 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package view contains the frontend for the client application.
package view

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)

// View is a text-based graphical user interface.
type View struct {
	terminal  *tcell.Terminal
	container *container.Container

	gaugeText                *text.Text
	sendingRateGauge         *gauge.Gauge
	historicalSendRateClient *linechart.LineChart
	historicalSendRateServer *linechart.LineChart
	console                  *text.Text

	mu sync.Mutex
	// maxSendRate is the all-time max send rate
	maxSendRate int
}

func New() (*View, error) {
	t, err := tcell.New()
	if err != nil {
		return nil, err
	}

	topWidget, err := text.New()
	if err != nil {
		return nil, err
	}
	if err := topWidget.Write("top widget"); err != nil {
		return nil, err
	}

	bottomWidget, err := text.New()
	if err != nil {
		return nil, err
	}
	if err := bottomWidget.Write("bottom widget"); err != nil {
		return nil, err
	}

	currentSendingRate, err := gauge.New(
		gauge.Height(1),
		gauge.Color(cell.ColorCyan),
		gauge.HideTextProgress(),
	)
	if err != nil {
		return nil, err
	}
	currentSendingRate.Absolute(0, 1)

	historicalSendingRateClient, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
	)
	if err != nil {
		return nil, err
	}

	historicalSendingRateServer, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
	)
	if err != nil {
		return nil, err
	}

	gaugeText, err := text.New()
	if err != nil {
		return nil, err
	}

	console, err := text.New(text.RollContent(), text.WrapAtWords())
	if err != nil {
		return nil, err
	}

	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("SCION Bandwidth Reservation Test Tool (Press Q to quit)"),
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.BorderTitle("Current sending rate"),
				container.SplitHorizontal(
					container.Top(
						container.SplitVertical(
							container.Left(),
							container.Right(
								container.PlaceWidget(gaugeText),
							),
						),
					),
					container.Bottom(
						// container.Border(linestyle.Light),
						container.PlaceWidget(currentSendingRate),
					),
					container.SplitFixed(1),
				),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.BorderTitle("Historical sending rate - client (mbps)"),
						container.PlaceWidget(historicalSendingRateClient),
					),
					container.Bottom(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.BorderTitle("Historical receiving rate - server (mbps)"),
								container.PlaceWidget(historicalSendingRateServer),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.BorderTitle("Console output"),
								container.PlaceWidget(console),
							),
							container.SplitPercent(66),
						),
					),
					container.SplitPercent(40),
				),
			),
			container.SplitFixed(4),
		),
	)
	if err != nil {
		return nil, err
	}

	return &View{
		terminal:  t,
		container: c,

		gaugeText:                gaugeText,
		sendingRateGauge:         currentSendingRate,
		historicalSendRateClient: historicalSendingRateClient,
		historicalSendRateServer: historicalSendingRateServer,
		console:                  console,
	}, nil
}

func (v *View) Run(ctx context.Context) error {
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' || k.Key == keyboard.KeyCtrlC {
			cancel()
		}
	}

	return termdash.Run(subCtx, v.terminal, v.container, termdash.KeyboardSubscriber(quitter))
}

// UpdateGauge reports the current sending rate (in bits per second). The
// rendered value is relative to the max seen sending rate.
func (v *View) UpdateGauge(bps int) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if bps > v.maxSendRate {
		v.maxSendRate = bps
	}
	v.sendingRateGauge.Absolute(bps, v.maxSendRate)
	v.gaugeText.Write(
		RenderBandwidth(bps, v.maxSendRate),
		text.WriteReplace(),
		text.WriteCellOpts(
			cell.FgColor(cell.ColorCyan),
		),
	)
}

// UpdateClientGraph updates the historical sending rate of the client graph.
func (v *View) UpdateClientGraph(values []float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.historicalSendRateClient.Series(
		"Send rate",
		values,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorCyan)),
		linechart.SeriesXLabels(map[int]string{
			0: "time",
		}),
	)
}

// UpdateServerGraph updates the historical sending rate of the server graph.
func (v *View) UpdateServerGraph(values []float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.historicalSendRateServer.Series(
		"Send rate",
		values,
		linechart.SeriesCellOpts(cell.FgColor(cell.ColorCyan)),
		linechart.SeriesXLabels(map[int]string{
			0: "time",
		}),
	)
}

// Print prints str in the console output widget.
func (v *View) Print(str string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.console.Write(str)
}

func (v *View) Close() {
	v.terminal.Close()
}

func RenderBandwidth(current, max int) string {
	currentF := float64(current)
	maxF := float64(max)

	measure := "bps"
	if maxF > 1000 {
		maxF /= 1000
		currentF /= 1000
		measure = "kbps"
	}
	if maxF > 1000 {
		maxF /= 1000
		currentF /= 1000
		measure = "mbps"
	}
	if maxF > 1000 {
		maxF /= 1000
		currentF /= 1000
		measure = "gpbs"
	}
	if measure == "bps" {
		return fmt.Sprintf("%.0f/%.0f %v", currentF, maxF, measure)
	}
	return fmt.Sprintf("%.2f/%.2f %v", currentF, maxF, measure)
}

// SeriesBuilder can be used to constructs inputs for a View timeseries.
type SeriesBuilder struct {
	mu sync.Mutex

	Graph      GraphUpdater
	WindowSize int

	series []float64
	index  int

	lastValue float64
}

func (b *SeriesBuilder) Update(value float64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If no data, just copy the last seen value so we have a contiguous plot line.
	if math.IsNaN(value) {
		value = b.lastValue
	}

	// XXX(scrye): This leaks memory, but because this is not a long-running app it should
	// not be a problem.
	if b.series == nil {
		b.series = make([]float64, b.WindowSize)
		b.index = b.WindowSize
	}
	b.series = append(b.series, value)
	b.index++
	b.Graph.UpdateGraph(b.series[b.index-b.WindowSize : b.index])
}

type GraphUpdater interface {
	UpdateGraph(values []float64)
}

type GraphUpdaterFunc func(values []float64)

func (f GraphUpdaterFunc) UpdateGraph(values []float64) {
	f(values)
}
