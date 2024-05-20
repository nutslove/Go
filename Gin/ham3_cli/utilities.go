package main

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

func Spinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[11], 200*time.Millisecond)
	s.Prefix = color.New(color.FgGreen).Sprint(message)
	s.Color("fgHiGreen")

	return s
}
