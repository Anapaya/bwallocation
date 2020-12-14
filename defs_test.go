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

package bwallocation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anapaya/bwallocation"
	"github.com/scionproto/scion/go/lib/xtest"
)

func TestLinkUnmarshalText(t *testing.T) {
	testCases := map[string]struct {
		Input        string
		Expected     bwallocation.Link
		ErrAssertion assert.ErrorAssertionFunc
	}{
		"valid": {
			Input: "1-ff00:0:110#1",
			Expected: bwallocation.Link{
				IA:        xtest.MustParseIA("1-ff00:0:110"),
				Interface: 1,
			},
			ErrAssertion: assert.NoError,
		},
		"invalid AS": {
			Input:        "1-ff00::0:110#1",
			ErrAssertion: assert.Error,
		},
		"invalid interface": {
			Input:        "1-ff00:0:110#-1",
			ErrAssertion: assert.Error,
		},
		"garbage": {
			Input:        "hello",
			ErrAssertion: assert.Error,
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			var l bwallocation.Link
			err := l.UnmarshalText([]byte(tc.Input))
			tc.ErrAssertion(t, err)
			assert.Equal(t, tc.Expected, l)
		})
	}
}

func TestLinkMarshalText(t *testing.T) {
	testCases := map[string]struct {
		Link     bwallocation.Link
		Expected string
	}{
		"valid": {
			Link: bwallocation.Link{
				IA:        xtest.MustParseIA("1-ff00:0:110"),
				Interface: 1,
			},
			Expected: "1-ff00:0:110#1",
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			raw, err := tc.Link.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tc.Expected, string(raw))
		})
	}
}

func TestBandwidthUnmarshalText(t *testing.T) {
	testCases := map[string]struct {
		Input        string
		Expected     bwallocation.Bandwidth
		ErrAssertion assert.ErrorAssertionFunc
	}{
		"2bps": {
			Input:        "2bps",
			Expected:     2,
			ErrAssertion: assert.NoError,
		},
		"1000bps": {
			Input:        "1000bps",
			Expected:     1000,
			ErrAssertion: assert.NoError,
		},
		"1kbps": {
			Input:        "1kbps",
			Expected:     1000,
			ErrAssertion: assert.NoError,
		},
		"1000kbps": {
			Input:        "1000kbps",
			Expected:     1000000,
			ErrAssertion: assert.NoError,
		},
		"1mbps": {
			Input:        "1mbps",
			Expected:     1000000,
			ErrAssertion: assert.NoError,
		},
		"1000mps": {
			Input:        "1000mbps",
			Expected:     1000000000,
			ErrAssertion: assert.NoError,
		},
		"1gbps": {
			Input:        "1gbps",
			Expected:     1000000000,
			ErrAssertion: assert.NoError,
		},
		"15gbps": {
			Input:        "15gbps",
			Expected:     15000000000,
			ErrAssertion: assert.NoError,
		},
		"float": {
			Input:        "1.5kbps",
			ErrAssertion: assert.Error,
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			var b bwallocation.Bandwidth
			err := b.UnmarshalText([]byte(tc.Input))
			tc.ErrAssertion(t, err)
			assert.Equal(t, int64(tc.Expected), int64(b))
		})
	}
}

func TestBandwidthMarshalText(t *testing.T) {
	testCases := map[string]struct {
		Bandwidth bwallocation.Bandwidth
		Expected  string
	}{
		"2bps": {
			Bandwidth: 2,
			Expected:  "2bps",
		},
		"1500bps": {
			Bandwidth: 1500,
			Expected:  "1500bps",
		},
		"1kbps": {
			Bandwidth: 1000,
			Expected:  "1kbps",
		},
		"1500kbps": {
			Bandwidth: 1500000,
			Expected:  "1500kbps",
		},
		"1mbps": {
			Bandwidth: 1000000,
			Expected:  "1mbps",
		},
		"1500mps": {
			Bandwidth: 1500000000,
			Expected:  "1500mbps",
		},
		"1gbps": {
			Bandwidth: 1000000000,
			Expected:  "1gbps",
		},
		"15gbps": {
			Bandwidth: 15000000000,
			Expected:  "15gbps",
		},
	}
	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			raw, err := tc.Bandwidth.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tc.Expected, string(raw))
		})
	}
}
