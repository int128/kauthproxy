package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/sync/errgroup"
)

func init() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
}

func main() {
	ctx := context.Background()
	eg, ctx := errgroup.WithContext(ctx)
	chInterrupt := make(chan struct{})
	eg.Go(func() error {
		defer close(chInterrupt)
		return runBrowser(ctx)
	})
	eg.Go(func() error {
		return runKauthproxy(chInterrupt)
	})
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}

func runKauthproxy(chInterrupt <-chan struct{}) error {
	c := exec.Command("output/kauthproxy",
		"--namespace=kubernetes-dashboard",
		"--user=tester",
		"--skip-open-browser",
		"https://kubernetes-dashboard-kong-proxy.svc",
	)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Start(); err != nil {
		return fmt.Errorf("could not start a process: %w", err)
	}
	log.Printf("started %s", c.String())
	<-chInterrupt
	if err := c.Process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("could not send SIGINT to the process: %w", err)
	}
	if err := c.Wait(); err != nil {
		return fmt.Errorf("could not wait for the process: %w", err)
	}
	return nil
}

func runBrowser(ctx context.Context) error {
	execOpts := chromedp.DefaultExecAllocatorOptions[:]
	execOpts = append(execOpts, chromedp.NoSandbox)
	ctx, cancel := chromedp.NewExecAllocator(ctx, execOpts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(2048, 1152),
		// open the page of pod list
		navigate("http://localhost:18000/#/pod?namespace=kube-system"),
		// wait for a link on the page
		chromedp.WaitReady(`a[href^='#/pod/kube-system']`, chromedp.ByQuery),
		takeScreenshot("output/screenshot.png"),
	)
	if err != nil {
		return fmt.Errorf("could not run the browser: %w", err)
	}
	return nil
}

// navigate to the URL and retry on network errors
func navigate(urlstr string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for {
			var l string
			if err := chromedp.Run(ctx,
				chromedp.Navigate(urlstr),
				chromedp.Location(&l),
			); err != nil {
				return err
			}
			log.Printf("opened %s", l)
			if strings.HasPrefix(l, "http://") {
				return nil
			}
			if err := chromedp.Sleep(1 * time.Second).Do(ctx); err != nil {
				return err
			}
		}
	})
}

// https://github.com/chromedp/examples/blob/master/screenshot/main.go
func takeScreenshot(name string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		var b []byte
		if err := chromedp.FullScreenshot(&b, 90).Do(ctx); err != nil {
			return fmt.Errorf("could not capture a screenshot: %w", err)
		}
		if err := os.WriteFile(name, b, 0644); err != nil {
			return fmt.Errorf("could not write: %w", err)
		}
		log.Printf("saved screenshot to %s", name)
		return nil
	})
}
