/*
Copyright 2021 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package diagnose

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	. "knative.dev/kn-plugin-diag/pkg/utils"

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	n       string
	verbose string
)

// domainCmd represents the domain command
func NewServiceCmd(p *ConnectionConfig) *cobra.Command {
	var serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "kn-diag service",
		Long: `Query knative service details. For example
kn-diag service <ksvc-name>`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf(`'service' requires a input arguments for knative servie name.
For example: kn-diag service <ksvc-name> -ns <namespace>`)
			}
			return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ksvcName := args[0]
			Namespace := "default"
			if cmd.Flags().Changed("namespace") {
				Namespace = n
			}
			sc, err := NewServingConfiguration(ksvcName, Namespace, p)
			if err != nil {
				return err
			}
			err = dumpToTables(sc, strings.ToLower(verbose))
			if err != nil {
				return err
			}
			return nil
		},
	}

	serviceCmd.Flags().StringVarP(&n, "namespace", "n", "", "the target namespace")
	serviceCmd.Flags().StringVarP(&verbose, "verbose", "", "", "enable verbose output. Supported value: keyinfo")
	return serviceCmd
}

func dumpToTables(sc *ServingConfiguration, verbose string) error {

	var table Table

	switch strings.ToLower(verbose) {
	case "keyinfo":
		table = NewTable(os.Stdout, []string{"Resource Type", "Name", "KeyInfo"})
	default:
		table = NewTable(os.Stdout, []string{"Resource Type", "Name", "Created At", "Status.Condition"})
	}

	defer table.Print()

	err := sc.deepFirstRetrieveObjects(sc.objectRoot, 0, table, verbose)
	if err != nil {
		return err
	}

	return nil

}
