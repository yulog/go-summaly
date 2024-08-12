package server

import (
	_ "embed"
	"fmt"

	"github.com/fatih/color"
)

// https://patorjk.com/software/taag/#p=display&f=Small%20Slant&t=GoSummaly
//
//go:embed banner.txt
var banner string

func PrintBanner(version string) {
	fmt.Printf(banner+"\n", color.New(color.FgRed).SprintFunc()("v"+version))
	fmt.Println("Yet another summaly proxy, written in Go")
}
