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
