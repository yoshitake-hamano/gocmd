
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yoshitake-hamano/gocmd/line"
	"github.com/chromedp/chromedp"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func mindraURL(lat, lon float64) string {
	return fmt.Sprintf("https://9db.jp/dqwalk/map#%f,%f,14", lat, lon)
}

func scrape(lat, lon float64, headless bool) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	if headless == false {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
			chromedp.Flag("disable-gpu", false),
			chromedp.Flag("enable-automation", false),
			chromedp.Flag("disable-extensions", false),
			chromedp.Flag("hide-scrollbars", false),
			chromedp.Flag("mute-audio", false),
		)
	
		allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	
		// create chrome instance
		ctx, cancel = chromedp.NewContext(
			allocCtx,
			chromedp.WithLogf(log.Printf),
		)
	}
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

		chromedp.Screenshot(`#map`, &imageBuf, chromedp.NodeVisible, chromedp.ByID),
	)
	if err != nil {
		return nil, err
	}

	return imageBuf, err
}

func main() {

	img, err := scrape(MINDRA_LAT, MINDRA_LON, true)
	check(err)

	url := mindraURL(MINDRA_LAT, MINDRA_LON)
	err = line.NotifyLineImage(bytes.NewReader(img), "map.jpg", url, LINE_NOTIFY_TOKEN)
	check(err)
}
