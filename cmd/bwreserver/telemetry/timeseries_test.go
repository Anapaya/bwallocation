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

package telemetry_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/anapaya/bwallocation/cmd/bwreserver/telemetry"
)

func TestTimeSeries(t *testing.T) {
	start := time.Now()

	ts := telemetry.NewTimeSeries(start)

	ts.Add(telemetry.Entry{Timestamp: start.Add(2 * time.Second), Value: 5})

	// FIXME(scrye): It just so happens that these floats test well against equality. Ideally,
	// float equality comparisons should be done with error allowance.

	assert.Equal(t, float64(2.5), ts.Difference(start, start.Add(time.Second)))
	assert.Equal(t, float64(5), ts.Difference(start, start.Add(2*time.Second)))
	assert.True(t, math.IsNaN(ts.Difference(start, start.Add(4*time.Second))))
	assert.True(t, math.IsNaN(ts.Difference(start.Add(3*time.Second), start.Add(4*time.Second))))

	ts.Add(telemetry.Entry{Timestamp: start.Add(4 * time.Second), Value: 15})
	assert.Equal(t, float64(15), ts.Difference(start, start.Add(4*time.Second)))
	assert.Equal(t, float64(7.5), ts.Difference(start.Add(time.Second), start.Add(3*time.Second)))
}

func TestEmptyPerSecond(t *testing.T) {
	ts := telemetry.NewTimeSeries(time.Now())

	windowStart := time.Now().Add(-2 * time.Second)
	windowEnd := time.Now().Add(-time.Second)

	assert.Equal(t, float64(0), telemetry.PerSecond(ts, windowStart, windowEnd))
}

func TestPerSecond(t *testing.T) {
	start := time.Now()

	ts := telemetry.NewTimeSeries(start)
	ts.Add(telemetry.Entry{Timestamp: start.Add(10 * time.Second), Value: 10})
	ts.Add(telemetry.Entry{Timestamp: start.Add(20 * time.Second), Value: 20})

	// 20 increase in 20 seconds is 1 per second
	assert.Equal(t, float64(1),
		telemetry.PerSecond(ts, start.Add(5*time.Second), start.Add(15*time.Second)))
}
