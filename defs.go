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

package bwallocation

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/serrors"
)

// Link represents a link from a leaf AS to the provider.
type Link struct {
	IA        addr.IA
	Interface uint64
}

func (l *Link) UnmarshalText(text []byte) error {
	parts := bytes.Split(text, []byte("#"))
	if len(parts) != 2 {
		return serrors.New("invalid link", "input", string(text))
	}
	ia, err := addr.IAFromString(string(parts[0]))
	if err != nil {
		return serrors.WrapStr("invalid link: failed to parse ISD-AS", err)
	}
	intf, err := strconv.ParseUint(string(parts[1]), 10, 64)
	if err != nil {
		return serrors.WrapStr("invalid link: failed to parse interface", err)
	}
	*l = Link{
		IA:        ia,
		Interface: uint64(intf),
	}
	return nil
}

func (l Link) MarshalText() (text []byte, err error) {
	return []byte(l.String()), nil
}

func (l Link) String() string {
	return fmt.Sprintf("%s#%d", l.IA, l.Interface)
}

// Bandwidth is the bandwidth in bps
type Bandwidth uint64

func (b *Bandwidth) UnmarshalText(text []byte) error {
	factor := uint64(1)
	raw := string(text)
	switch {
	case strings.HasSuffix(raw, "gbps"):
		raw, factor = strings.TrimSuffix(raw, "gbps"), 1e9
	case strings.HasSuffix(raw, "mbps"):
		raw, factor = strings.TrimSuffix(raw, "mbps"), 1e6
	case strings.HasSuffix(raw, "kbps"):
		raw, factor = strings.TrimSuffix(raw, "kbps"), 1e3
	case strings.HasSuffix(raw, "bps"):
		raw, factor = strings.TrimSuffix(raw, "bps"), 1
	}
	val, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return serrors.WrapStr("parsing bandwidth", err, "input", string(text))
	}
	*b = Bandwidth(val * factor)
	return nil
}

func (b Bandwidth) MarshalText() (text []byte, err error) {
	return []byte(b.String()), nil
}

func (b Bandwidth) String() string {
	switch {
	case b >= 1e9 && b%1e9 == 0:
		return fmt.Sprintf("%dgbps", b/1e9)
	case b >= 1e6 && b%1e6 == 0:
		return fmt.Sprintf("%dmbps", b/1e6)
	case b >= 1e3 && b%1e3 == 0:
		return fmt.Sprintf("%dkbps", b/1e3)
	default:
		return fmt.Sprintf("%dbps", b)
	}
}
