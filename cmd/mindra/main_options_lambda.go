// +build lambda

package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/chromedp/chromedp"
)

func appendExecAllocatorOptions(opts []chromedp.ExecAllocatorOption) []chromedp.ExecAllocatorOption {
	// https://stackoverflow.com/questions/49402887/errorbrowser-gpu-channel-host-factory-cc121-failed-to-launch-gpu-process
	opts = append(opts,
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.ExecPath("/opt/headless-chromium/headless-chromium"),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("homedir", "/tmp"),
		chromedp.Flag("data-path", "/tmp/data-path"),
		chromedp.Flag("disk-cache-dir", "/tmp/cache-dir"),
		// https://github.com/adieuadieu/serverless-chrome/issues/108
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("enable-webgl", true),
		chromedp.Flag("use-gl", "osmesa"),
	)
	return opts
}

func HandleRequest(ctx context.Context, e MyEvent) (string, error) {
	param := fmt.Sprintf("lat=%f, lon=%f", e.Lat, e.Lon)
	err := mainImpl(e.Lat, e.Lon)
	return param, err
}

func main() {
	lambda.Start(HandleRequest)
}
