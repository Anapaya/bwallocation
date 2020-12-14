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

package reservation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/anapaya/bwallocation/reservation"
	"github.com/scionproto/scion/go/lib/slayers/path"
	"github.com/scionproto/scion/go/lib/slayers/path/scion"
)

var rawPath = []byte("\x00\x00\x00\x01\xef\xef\xef\xef\x00\x60\x01\x02\x03\x04\x05\x06" +
	"\x00\x00\x20\x80\x00\x00\x01\x11\x00\x00\x01\x00\x01\x00\x02\x22\x00\x00\x01\x00\x00\x3f\x00" +
	"\x01\x00\x00\x01\x02\x03\x04\x05\x06\x00\x3f\x00\x03\x00\x02\x01\x02\x03\x04\x05\x06\x00\x3f" +
	"\x00\x00\x00\x02\x01\x02\x03\x04\x05\x06\x00\x3f\x00\x01\x00\x00\x01\x02\x03\x04\x05\x06")

var testPath = &reservation.Path{
	MetaHdr: reservation.MetaHdr{
		ID:        1,
		ExpTime:   0xefefefef,
		RateLimit: 96,
		MAC:       []uint8{1, 2, 3, 4, 5, 6},
	},
	Decoded: scion.Decoded{
		Base: scion.Base{
			PathMeta: scion.MetaHdr{
				CurrINF: 0,
				CurrHF:  0,
				SegLen:  [3]uint8{2, 2, 0},
			},

			NumINF:  2,
			NumHops: 4,
		},
		InfoFields: testInfoFields,
		HopFields:  testHopFields,
	},
}

var testInfoFields = []*path.InfoField{
	{
		Peer:      false,
		ConsDir:   false,
		SegID:     0x111,
		Timestamp: 0x100,
	},
	{
		Peer:      false,
		ConsDir:   true,
		SegID:     0x222,
		Timestamp: 0x100,
	},
}

var testHopFields = []*path.HopField{
	{
		ExpTime:     63,
		ConsIngress: 1,
		ConsEgress:  0,
		Mac:         []byte{1, 2, 3, 4, 5, 6},
	},
	{
		ExpTime:     63,
		ConsIngress: 3,
		ConsEgress:  2,
		Mac:         []byte{1, 2, 3, 4, 5, 6},
	},
	{
		ExpTime:     63,
		ConsIngress: 0,
		ConsEgress:  2,
		Mac:         []byte{1, 2, 3, 4, 5, 6},
	},
	{
		ExpTime:     63,
		ConsIngress: 1,
		ConsEgress:  0,
		Mac:         []byte{1, 2, 3, 4, 5, 6},
	},
}

func TestSerialize(t *testing.T) {
	b := make([]byte, testPath.Len())
	assert.NoError(t, testPath.SerializeTo(b))
	assert.Equal(t, rawPath, b)
}

func TestDecodeFromBytes(t *testing.T) {
	s := &reservation.Path{}
	assert.NoError(t, s.DecodeFromBytes(rawPath))
	assert.Equal(t, testPath, s)
}

func TestSerializeDecode(t *testing.T) {
	b := make([]byte, testPath.Len())
	assert.NoError(t, testPath.SerializeTo(b))
	s := &reservation.Path{}
	assert.NoError(t, s.DecodeFromBytes(b))
	assert.Equal(t, testPath, s)
}

func TestReverse(t *testing.T) {
	revPath, err := testPath.Reverse()
	assert.NoError(t, err)
	s := &scion.Decoded{}
	assert.NoError(t, s.DecodeFromBytes(rawPath[reservation.MetaLen:]))
	sRevPath, err := s.Reverse()
	assert.NoError(t, err)
	assert.Equal(t, sRevPath, revPath)
}
