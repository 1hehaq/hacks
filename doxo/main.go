package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	dollar = "\033[32m$\033[32m"
	arrow  = "\033[32m→\033[32m"
)

func main() {
	var (
		webhook = flag.String("hook", "", "Discord webhook URL")
		text    = flag.String("text", "", "Text message to send")
		file    = flag.String("file", "", "File path to send")
		tts     = flag.Bool("tts", false, "Enable text-to-speech")
		silent  = flag.Bool("silent", false, "Suppress output")
		ss      = flag.Bool("ss", false, "Capture and send terminal screenshot")
		cmd     = flag.String("cmd", "", "Command to execute and screenshot")
		nocmd   = flag.Bool("nocmd", false, "Hide command in screenshot")
		help    = flag.Bool("help", false, "Show help")
	)

	flag.Usage = printHelp
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *webhook == "" {
		*webhook = os.Getenv("DOXO_WEBHOOK")
	}
	if *webhook == "" {
		exitError("Webhook not provided", *silent)
	}

	if *ss {
		handleScreenshot(*webhook, *text, *cmd, *nocmd, *tts, *silent)
		return
	}

	content := *text
	if content == "" && *file == "" {
		if stdin, err := io.ReadAll(os.Stdin); err == nil {
			content = string(stdin)
		}
	}

	if content == "" && *file == "" {
		exitError("No content provided", *silent)
	}

	if *file != "" {
		sendFile(*webhook, content, *file, *tts, *silent)
	} else {
		sendText(*webhook, content, *tts, *silent)
	}
}

func handleScreenshot(webhook, text, cmdStr string, nocmd, tts, silent bool) {
	if cmdStr == "" {
		exitError("No command provided for screenshot", silent)
	}

	tempFile := createTempFile(silent)
	defer os.Remove(tempFile)

	if !runTermshot(tempFile, cmdStr, nocmd, silent) {
		exitError("Screenshot capture failed", silent)
	}

	sendFile(webhook, text, tempFile, tts, silent)
}

func createTempFile(silent bool) string {
	f, err := os.CreateTemp("", "doxo-*.png")
	if err != nil {
		exitError("Temp file creation failed", silent)
	}
	defer f.Close()
	return f.Name()
}

func runTermshot(filename, cmdStr string, nocmd, silent bool) bool {
	args := []string{"-f", filename, "-s", "--no-shadow"}
	if !nocmd {
		args = append(args, "-c")
	}
	args = append(args, "--", cmdStr)

	cmd := exec.Command("termshot", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run() == nil
}

func sendText(webhook, content string, tts, silent bool) {
	payload, _ := json.Marshal(map[string]interface{}{
		"content": content,
		"tts":     tts,
	})

	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(payload))
	if err != nil || resp.StatusCode > 299 {
		exitError("Message send failed", silent)
	}
}

func sendFile(webhook, content, path string, tts, silent bool) {
	file, err := os.Open(path)
	if err != nil {
		exitError("File open failed", silent)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, _ := writer.CreateFormFile("file", filepath.Base(path))
	io.Copy(part, file)

	jsonData, _ := json.Marshal(map[string]interface{}{
		"content": content,
		"tts":     tts,
	})
	writer.WriteField("payload_json", string(jsonData))
	writer.Close()

	resp, err := http.Post(webhook, writer.FormDataContentType(), body)
	if err != nil || resp.StatusCode > 299 {
		exitError("File send failed", silent)
	}
}

func printHelp() {
	fmt.Printf("\n%s\t ┓%s\n", bold+red, reset)
	fmt.Printf("%s\t┏┫┏┓┓┏┏┓ %s{v0.2}\n", bold+red, reset)
	fmt.Printf("%s\t┗┻┗┛┛┗┗┛%s\n\n", bold+red, reset)
	
	fmt.Printf("%sUSAGE%s:\n", bold+yellow, reset)
	fmt.Printf("%s %sdoxo [flags]\n", dollar, reset)
	fmt.Printf("%s %s[piped input] | doxo [flags]\n", dollar, reset)
	fmt.Printf("%s %sdoxo -ss -cmd \"command\"\n\n", dollar, reset)

	fmt.Printf("%sFLAGS%s:\n", bold+yellow, reset)
	fmt.Printf("%s %s-hook   %sURL%s    discord webhook URL\n", arrow, reset, yellow, reset)
	fmt.Printf("%s %s-text   %sTEXT%s   message content\n", arrow, reset, yellow, reset)
	fmt.Printf("%s %s-file   %sPATH%s   file to upload\n", arrow, reset, yellow, reset)
	fmt.Printf("%s %s-cmd    %sCMD%s    command to execute\n", arrow, reset, yellow, reset)
	fmt.Printf("%s %s-ss            capture screenshot\n", arrow, reset)
	fmt.Printf("%s %s-nocmd         hide command in screenshot\n", arrow, reset)
	fmt.Printf("%s %s-tts           enable text-to-speech\n", arrow, reset)
	fmt.Printf("%s %s-silent        suppress output\n", arrow, reset)
	fmt.Printf("%s %s-help          show help\n\n", arrow, reset)
	
	fmt.Printf("%sNOTES:%s\n", bold+green, reset)
	fmt.Printf("  • Set %sDOXO_WEBHOOK%s environment variable\n", bold, reset)
	fmt.Printf("  • Screenshots require -cmd flag\n\n")
}

func exitError(msg string, silent bool) {
	if !silent {
		fmt.Printf("%s✗ %s%s\n", red, msg, reset)
	}
	os.Exit(1)
}
