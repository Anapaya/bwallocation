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

// Package bwallocation implements the control plane parts of the bandwidth
// allocation system for star-topologies.
//
// At the core is the Manager. It manages bandwidth reservation subscriptions
// that can be used to send traffic with reserved bandwidth. Client applications
// that want to use bandwidth reservations to send traffic initialize the
// Manager and subscribe for a subscription with a set of parameters. The
// manager takes care of establishing and renewing the reservation for all of
// its entire lifetime.
//
// Subscriptions simply return the latest reservation token that the client
// application should use for its communication. After the client is done using
// the reservation, it should close the subscription to abort unnecessary
// renewal. This is especially important for subscriptions with unbounded
// expiration time.
//
// To use the reservation for sending traffic, simply use the Conn wrapper and
// pass the subscription to it.
package bwallocation
