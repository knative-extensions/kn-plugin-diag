package utils

import (
	"fmt"

	"github.com/fatih/color"
)

func SayOK() {
	c := color.New(color.FgGreen).Add(color.Bold)
	c.Println("OK" + "\n")
}

func SayFailed() {
	c := color.New(color.FgRed).Add(color.Bold)
	c.Println("FAILED" + "\n")
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
