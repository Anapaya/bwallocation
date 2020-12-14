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

package reservation

import (
	"bytes"
	"fmt"
	"hash"

	"github.com/scionproto/scion/go/lib/serrors"
	"github.com/scionproto/scion/go/lib/slayers"
)

// MAC calculates the Reservation MAC.
func MAC(h hash.Hash, scion *slayers.SCION, path *Path) []byte {
	h.Reset()
	input := MACInput(scion, path)
	// Write must not return an error: https://godoc.org/hash#Hash
	if _, err := h.Write(input); err != nil {
		panic(err)
	}
	return h.Sum(nil)[:6]
}

// VerifyMAC verifies that the MAC in the Reservation header is correct, i.e.,
// matches the value calculated with MAC(h, scion, path). Returns nil if the
// MACs match, error otherwise.
func VerifyMAC(h hash.Hash, scion *slayers.SCION, path *Path) error {
	expectedMac := MAC(h, scion, path)
	if !bytes.Equal(path.MAC, expectedMac) {
		return serrors.New("MAC",
			"expected", fmt.Sprintf("%x", expectedMac),
			"actual", fmt.Sprintf("%x", path.MAC))
	}
	return nil
}

// MACInput returns the MAC input data block with the following layout:
//
//   0                   1                   2                   3
//   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |            DstISD             |                               |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+                               +
//  |                             DstAS                             |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |            SrcISD             |                               |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+                               +
//  |                             SrcAS                             |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |                           DstHost (var)                       |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |                           SrcHost (var)                       |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |                              ID                               |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |                            ExpTime                            |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |           RateLimit           |                               |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+                               +
//  |                               0                               |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//  |                        SCION Path (var)                       |
//  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
func MACInput(scion *slayers.SCION, path *Path) []byte {
	addrLen := scion.AddrHdrLen()
	inputLen := addrLen + path.Len()
	input := make([]byte, inputLen)

	/* Save path MAC */
	mac := path.MAC
	path.MAC = nil

	offset := 0
	err := scion.SerializeAddrHdr(input[offset : offset+addrLen])
	if err != nil {
		panic(err)
	}
	offset += addrLen
	err = path.SerializeTo(input[offset:])
	if err != nil {
		panic(err)
	}

	/* Restore path MAC */
	path.MAC = mac

	return input
}
