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

// Package reservation defines the dataplane representation of the bandwidth
// reservation token in the bandwidth allocation system for star topologies.
package reservation

import (
	"encoding/binary"

	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/slayers/path"
	"github.com/scionproto/scion/go/lib/slayers/path/scion"
)

const (
	// PathType is the path type unique value of the reservation path.
	PathType path.Type = 250
	// MetaLen is the length of the PathMetaHeader.
	MetaLen = 16
)

// RegisterPath registers the path type with the path package. It must be
// invoked to activate the reservation path type.
func RegisterPath() {
	path.RegisterPath(path.Metadata{
		Type: PathType,
		Desc: "Reservation",
		New: func() path.Path {
			return &Path{}
		},
	})
}

// Path encodes a reservation path. A reservation path is essentially a SCION
// path with an additional path meta header.
type Path struct {
	MetaHdr
	scion.Decoded
}

// DecodeFromBytes populates the fields from a raw buffer.
// The buffer must be of length >= MetaLen.
func (p *Path) DecodeFromBytes(raw []byte) error {
	if err := p.MetaHdr.DecodeFromBytes(raw); err != nil {
		return err
	}
	if err := p.Decoded.DecodeFromBytes(raw[MetaLen:]); err != nil {
		return err
	}
	return nil
}

// SerializeTo writes the fields into the provided buffer.
// The buffer must be of length >= MetaLen.
func (p *Path) SerializeTo(b []byte) error {
	if err := p.MetaHdr.SerializeTo(b); err != nil {
		return err
	}
	if err := p.Decoded.SerializeTo(b[MetaLen:]); err != nil {
		return err
	}
	return nil
}

// Reverse reverses a Reservation path. The reservation token is stripped from
// the returned regular SCION path.
func (p *Path) Reverse() (path.Path, error) {
	return p.Decoded.Reverse()
}

// Len returns the length of the path in bytes.
func (p *Path) Len() int {
	return MetaLen + p.Decoded.Len()
}

// Type returns the type of the path.
func (p *Path) Type() path.Type {
	return PathType
}

// MetaHdr is the PathMetaHdr of a Reservation (data-plane) path type.
//
//   0                   1                   2                   3
//   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |                              ID                               |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |                            ExpTime                            |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |           RateLimit           |                               |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+                               +
//  |                              Mac                              |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
type MetaHdr struct {
	// ID represents the unique ID of the reservation.
	ID uint32
	// ExpTime is the time that the reservation expires, expressed in Unix time as a
	// 4 byte integer with 1 second granularity.
	ExpTime uint32
	// RateLimit is the resevation rate limit with 1 Mbit/s granularity.
	RateLimit uint16
	// MAC is the 6-byte Message Authentication Code to authenticate the reservation.
	MAC []byte
}

// DecodeFromBytes populates the fields from a raw buffer.
// The buffer must be of length >= MetaLen.
func (m *MetaHdr) DecodeFromBytes(raw []byte) error {
	if len(raw) < MetaLen {
		return serrors.New("Reservation raw too short", "expected", MetaLen, "actual", len(raw))
	}
	m.ID = binary.BigEndian.Uint32(raw[0:4])
	m.ExpTime = binary.BigEndian.Uint32(raw[4:8])
	m.RateLimit = binary.BigEndian.Uint16(raw[8:10])
	m.MAC = append([]byte(nil), raw[10:MetaLen]...)

	return nil
}

// SerializeTo writes the fields into the provided buffer.
// The buffer must be of length >= MetaLen.
func (m *MetaHdr) SerializeTo(b []byte) error {
	if len(b) < MetaLen {
		return serrors.New("Reservation buffer too short", "expected", MetaLen, "actual", len(b))
	}
	binary.BigEndian.PutUint32(b[0:4], m.ID)
	binary.BigEndian.PutUint32(b[4:8], m.ExpTime)
	binary.BigEndian.PutUint16(b[8:10], m.RateLimit)
	copy(b[10:MetaLen], m.MAC)

	return nil
}
