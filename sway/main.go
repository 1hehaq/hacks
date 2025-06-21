package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	cyan := color.New(color.FgHiCyan, color.Bold)
	grey := color.New(color.FgHiBlack)
	
	msg := cyan.Sprint("INFO") + " "
	
	if target, ok := entry.Data["target"]; ok {
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
	
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		test(strings.TrimSpace(scanner.Text()))
	}
}

func test(url string) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	fields := logrus.Fields{"target": url}
	
	for _, method := range methods {
		if status := probe(url, method); status > 0 {
			fields[strings.ToLower(method)] = status
		}
	}
	
	if len(fields) > 1 {
		logrus.WithFields(fields).Info("")
	}
}

func probe(url, method string) int {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return 0
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	
	return resp.StatusCode
}
