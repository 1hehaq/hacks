package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	cyan := color.New(color.FgHiCyan, color.Bold)
	grey := color.New(color.FgHiBlack)

	msg := cyan.Sprint("INFO") + " "

	if target, ok := entry.Data["target"];
		ok {
		msg += grey.Sprint("target") + "=" + fmt.Sprintf("%v", target)

		methods := []string{"get", "post", "put", "delete", "patch", "options", "head"}
		for _, method := range methods {
			if status, exists := entry.Data[method]; exists {
				msg += " " + grey.Sprint(method) + "=" + fmt.Sprintf("\"%v\"", status)
			}
		}
	}

	return []byte(msg + "\n"), nil
}

func main() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&CustomFormatter{})
	logrus.SetOutput(os.Stdout) // Change output to stdout

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		test(strings.TrimSpace(scanner.Text()))
	}
}

func test(url string) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	fields := logrus.Fields{"target": url}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, method := range methods {
		wg.Add(1)
		go func(m string) {
			defer wg.Done()
			if status := probe(url, m); status > 0 {
				mu.Lock()
				fields[strings.ToLower(m)] = status
				mu.Unlock()
			}
		}(method)
	}

	wg.Wait()

	if len(fields) > 1 {
		logrus.WithFields(fields).Info("")
	}
}

func probe(url, method string) int {
	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0
	}

	req.Header.Set("User-Agent", "sway/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	return resp.StatusCode
}
