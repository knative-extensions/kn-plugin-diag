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

package utils

import (
	"fmt"

	"github.com/fatih/color"
)

func SayOK() {
	c := color.New(color.FgGreen).Add(color.Bold)
	c.Println("OK")
}

func SayFailed() {
	c := color.New(color.FgRed).Add(color.Bold)
	c.Println("FAILED")
}

func SayMessage(format string, args ...interface{}) {
	fmt.Printf(format+"%v\n", args...)
}

func SayWarningMessage(format string, args ...interface{}) {
	c := color.New(color.FgYellow).Add(color.Bold)
	c.Printf(format, args...)
}

func SayFailedMessage(format string, args ...interface{}) {
	c := color.New(color.FgRed).Add(color.Bold)
	c.Printf(format, args...)
}
