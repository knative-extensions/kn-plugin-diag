package main

import (
	"fmt"
	"os"
	// "strings"
	// "time"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"knative.dev/kn-plugin-diag/pkg/command/diagnose"
	"knative.dev/kn-plugin-diag/pkg/utils"
)

func main() {

	p := &utils.ConnectionConfig{}
	p.Initialize()

	rootCmd := &cobra.Command{
		Use:   "knative-diagnose",
		Short: "A CLI of to help with Diagnose Knative Resources",
		Long:  `A CLI of to help with Diagnose Knative Resources.`,
	}
	rootCmd.AddCommand(diagnose.NewServiceCmd(p))
	rootCmd.InitDefaultHelpCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Failed : ", err)
		os.Exit(1)
	}

}
