//go:build !lambda
// +build !lambda

package main

import (
	"fmt"

	"github.com/chromedp/chromedp"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func appendExecAllocatorOptions(opts []chromedp.ExecAllocatorOption) []chromedp.ExecAllocatorOption {
	return opts
}

func main() {
	fmt.Printf("lat=%f, lon=%f", MINDRA_LAT, MINDRA_LON)

	err := mainImpl(MINDRA_LAT, MINDRA_LON, LINE_NOTIFY_TOKEN)
	check(err)
}
