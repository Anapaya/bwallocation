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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/anapaya/bwallocation/cmd/bwreserver/client"
	"github.com/anapaya/bwallocation/cmd/bwreserver/server"
	"github.com/anapaya/bwallocation/reservation"
)

func main() {
	reservation.RegisterPath()

	executable := filepath.Base(os.Args[0])
	cmd := &cobra.Command{
		Use:   executable,
		Short: "Bandwidth Reservation Demo App",
		Long: fmt.Sprintf(`%s demonstrate bandwidth reservations.

In server mode, the application accepts new sessions from clients and measures
the receiving rate.

In client mode, the application attempts to connect to the server, reserve
bandwidth on the path, and send at the full rate of the reservation. The client
application periodically reports its measured sending rate, and the receiving
rate observed by the server.
`, executable),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
	}
	cmd.AddCommand(client.Cmd(cmd), server.Cmd(cmd))

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
