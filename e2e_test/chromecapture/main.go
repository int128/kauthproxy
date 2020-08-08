package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s URL", os.Args[0])
		return
	}
	url := os.Args[1]
	if err := runBrowser(context.Background(), url); err != nil {
		log.Fatalf("error: %s", err)
	}
}

func runBrowser(ctx context.Context, url string) error {
	execOpts := chromedp.DefaultExecAllocatorOptions[:]
	execOpts = append(execOpts, chromedp.NoSandbox)
	ctx, cancel := chromedp.NewExecAllocator(ctx, execOpts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := takeScreenshot(ctx, url); err != nil {
		return fmt.Errorf("could not run the browser: %w", err)
	}
	return nil
}

func takeScreenshot(ctx context.Context, url string) error {
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		logPageMetadata(),
		chromedp.Sleep(3*time.Second),
		logPageMetadata(),
		fullScreenshot(func(b []byte) error {
			if err := ioutil.WriteFile("fullScreenshot.png", b, 0644); err != nil {
				return fmt.Errorf("could not write: %w", err)
			}
			return nil
		}),
	)
	if err != nil {
		return err
	}
	return nil
}

// capture entire browser viewport:
// https://github.com/chromedp/examples/blob/master/screenshot/main.go
func fullScreenshot(f func([]byte) error) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// get layout metrics
		_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
		if err != nil {
			return fmt.Errorf("could not get layout metrics: %w", err)
		}

		width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

		// force viewport emulation
		err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
			WithScreenOrientation(&emulation.ScreenOrientation{
				Type:  emulation.OrientationTypePortraitPrimary,
				Angle: 0,
			}).
			Do(ctx)
		if err != nil {
			return fmt.Errorf("could not set viewport emulation: %w", err)
		}

		// capture screenshot
		b, err := page.CaptureScreenshot().
			WithClip(&page.Viewport{
				X:      contentSize.X,
				Y:      contentSize.Y,
				Width:  contentSize.Width,
				Height: contentSize.Height,
				Scale:  1,
			}).Do(ctx)
		if err != nil {
			return fmt.Errorf("could not capture a screenshot: %w", err)
		}
		return f(b)
	})
}

func logPageMetadata() chromedp.Action {
	var location string
	var title string
	return chromedp.Tasks{
		chromedp.Location(&location),
		chromedp.Title(&title),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Printf("location: %s [%s]", location, title)
			return nil
		}),
	}
}
