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

// Package telemetry adds a simple timeseries implementation for flexible computation of rates
// based on an append-only log of cumulative values. This is a toy implementation and not fit
// for production use.
package telemetry

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Entry is a timestamped value that can be added to a time series.
type Entry struct {
	Timestamp time.Time
	Value     float64
}

// TimeSeries is a simplified memory-only time series with monotonic timestamps.
// This is a toy implementation and not fit for production use. Use Prometheus
// instead.
//
// TimeSeries methods can be called concurrently from different goroutines.
type TimeSeries struct {
	mu sync.Mutex

	// entries is an append-only slice of entries added to the time series.
	entries []Entry

	maxTimestamp time.Time
	maxValue     float64
}

// NewTimeSeries initializes a new time series with a value of 0 at start time. Values before
// start are assumed to be all 0.
func NewTimeSeries(start time.Time) *TimeSeries {
	return &TimeSeries{
		entries: []Entry{
			{},                 // initialize sentinel value, makes lower bound searches simpler
			{Timestamp: start}, // initialize start, to simplify computation of rates with start
			// times before initial point of reference
		},
		maxTimestamp: start,
	}
}

// Add appends an entry to the time series. The entry is only added if its
// timestamp is newer than all entries in the series and the value is greater
// than all entries in the series. Otherwise, an error is returned. Add can be
// called concurrently from different goroutines.
func (ts *TimeSeries) Add(e Entry) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !e.Timestamp.After(ts.maxTimestamp) {
		return fmt.Errorf("timestamp %v is in the past", e.Timestamp)
	}
	if !(e.Value > ts.maxValue) {
		return fmt.Errorf("value %v is not monotonically increasing", e.Value)
	}
	ts.entries = append(ts.entries, e)
	return nil
}

// Difference computes the change in value between start and end. Start can be before the
// reference point set at time series creation. Values before the reference point are assumed
// to be 0.
//
// If start is greater or equal to end, the function panics. If end is past the last seen
// value observation, NaN is returned.
//
// Difference can be called concurrently from different goroutines.
func (ts *TimeSeries) Difference(start, end time.Time) float64 {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if !end.After(start) {
		panic(fmt.Sprintf("timestamp %v must be after %v", end, start))
	}

	startValue := ts.interpolateLocked(start)
	if math.IsNaN(startValue) {
		return math.NaN()
	}

	endValue := ts.interpolateLocked(end)
	if math.IsNaN(endValue) {
		return math.NaN()
	}

	return endValue - startValue
}

// interpolateLocked approximates the time series value at a point in time. It returns
// NaN if an approximation cannot be made due to unavailable data.
func (ts *TimeSeries) interpolateLocked(t time.Time) float64 {
	lowerIndex := ts.lowerLocked(t)
	if ts.lastIndexLocked(lowerIndex) {
		return math.NaN()
	}
	upperIndex := lowerIndex + 1

	lowerEntry := ts.entries[lowerIndex]
	upperEntry := ts.entries[upperIndex]

	durationFactor := float64(t.Sub(lowerEntry.Timestamp)) /
		float64(upperEntry.Timestamp.Sub(lowerEntry.Timestamp))
	return lowerEntry.Value + (upperEntry.Value-lowerEntry.Value)*durationFactor
}

// lastIndexLocked returns true if i is the last index in the time series.
func (ts *TimeSeries) lastIndexLocked(i int) bool {
	return i == len(ts.entries)-1
}

// lowerLocked finds the last entry that is less than t.
func (ts *TimeSeries) lowerLocked(t time.Time) int {
	// Search in reverse to optimize for queries biased towards the current time.
	for i := len(ts.entries) - 1; i >= 0; i-- {
		if ts.entries[i].Timestamp.Before(t) {
			return i
		}
	}
	panic("not reachable due to sentinel")
}

// PerSecond computes the rate of change between the start and end time of ts, when
// expressed in per second granularity. If the rate cannot be computed (e.g., due
// to unavailable data) NaN is returned.
func PerSecond(ts *TimeSeries, start, end time.Time) float64 {
	diff := ts.Difference(start, end)
	if math.IsNaN(diff) {
		return math.NaN()
	}
	f := end.Sub(start).Seconds()
	return diff / f
}
