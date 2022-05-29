package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/yoshitake-hamano/gocmd/line"
)

type MyEvent struct {
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Token string  `json:"token"`
}

func mindraURL(lat, lon float64) string {
	return fmt.Sprintf("https://9db.jp/dqwalk/map#%f,%f,14", lat, lon)
}

func scrape(lat, lon float64) ([]byte, error) {
	log.Printf("scrape(lat=%f, lon=%f)", lat, lon)

	opts := chromedp.DefaultExecAllocatorOptions[:]
	opts = append(opts,
		chromedp.CombinedOutput(os.Stdout),
		chromedp.Flag("vervose", true),
		chromedp.Flag("enable-logging", true),
		chromedp.Flag("log-level", "0"),
	)
	opts = appendExecAllocatorOptions(opts)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	var imageBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(mindraURL(lat, lon)),
		chromedp.Click(`.iziModal-button-close`, chromedp.ByQuery),
		// 祠OFF
		chromedp.Click(`li[data-id="h1"]`, chromedp.ByQuery),
		chromedp.Click(`li[data-id="h2"]`, chromedp.ByQuery),

		chromedp.Sleep(time.Second*5),
		chromedp.Screenshot(`#map`, &imageBuf, chromedp.NodeVisible, chromedp.ByID),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp.Run: %v", err)
	}

	return imageBuf, nil
}

func mainImpl(lat, lon float64, token string) error {
	fmt.Printf("lat=%f, lon=%f, token=%s", lat, lon, token)

	img, err := scrape(lat, lon)
	if err != nil {
		return fmt.Errorf("scrape: %v", err)
	}

	url := mindraURL(lat, lon)
	err = line.NotifyLineImage(bytes.NewReader(img), "map.jpg", url, token)
	if err != nil {
		fmt.Errorf("line.NotifyLineImage: %v", err)
	}
	return nil
}
