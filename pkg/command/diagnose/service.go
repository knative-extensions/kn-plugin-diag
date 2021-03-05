package diagnose

import (
	"fmt"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	n       string
	verbose string
)

//func NewServiceCmd(p *ConnectionConfig) *cobra.Command {
func NewServiceCmd() *cobra.Command {
	var serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "kn-diag service",
		Long: `Query knative service details. For example
knative-diagnose service <ksvc-name>\n`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf(`'service' requires a input arguments for knative servie name. For example: 
knative-diagnose service <ksvc-name> -ns <namespace>\n`)
			}

			return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Yes, it is ready to go!")
			return nil
		},
	}

	serviceCmd.Flags().StringVarP(&n, "namespace", "n", "", "the target namespace")
	serviceCmd.Flags().StringVarP(&verbose, "verbose", "", "", "enable verbose output")
	return serviceCmd
}
