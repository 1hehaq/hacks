package main

import (
	"bufio"
	"context"
	"flag"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/sirupsen/logrus"
)

type Config struct {
	threads int
	timeout int
	output  string
}

func main() {
	cfg := &Config{}
	flag.IntVar(&cfg.threads, "t", 10, "threads")
	flag.IntVar(&cfg.timeout, "timeout", 10, "timeout")
	flag.StringVar(&cfg.output, "o", "screenshots", "output")
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	})

	os.MkdirAll(cfg.output, 0755)
	run(cfg)
}

func run(cfg *Config) {
	var wg sync.WaitGroup
	jobs := make(chan string, 100)

	for i := 0; i < cfg.threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				takeScreenshot(target, cfg)
			}
		}()
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		target := scanner.Text()
		if isValidURL(target) {
			jobs <- target
		}
	}

	close(jobs)
	wg.Wait()
}

func takeScreenshot(target string, cfg *Config) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Duration(cfg.timeout)*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(target),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, _, contentSize, _, _, _, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).Do(ctx)
			if err != nil {
				return err
			}

			buf, err = page.CaptureScreenshot().
				WithClip(&page.Viewport{
					X: contentSize.X, Y: contentSize.Y,
					Width: contentSize.Width, Height: contentSize.Height, Scale: 1,
				}).Do(ctx)
			return err
		}),
	})

	filename := filepath.Join(cfg.output, sanitize(target)+".png")
	
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"target": target,
			"error":  err.Error(),
		}).Error("failed")
		return
	}

	if err := os.WriteFile(filename, buf, 0644); err != nil {
		logrus.WithFields(logrus.Fields{
			"target": target,
			"error":  err.Error(),
		}).Error("write failed")
		return
	}

	logrus.WithFields(logrus.Fields{
		"target": target,
		"output": filename,
	}).Info("screenshot")
}

func sanitize(s string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(s, "_")
}

func isValidURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}
