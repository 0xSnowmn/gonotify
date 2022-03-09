package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	flags "github.com/jessevdk/go-flags"
	_ "github.com/joho/godotenv"
)

// Setting Options
var Options struct {
	// Data that will send
	Data    string `short:"d" long:"data" description:"data sends to channel" required:"false"`
	Channel string `short:"w" long:"chan" description:"channel" required:"false"`
	// Concurrency For the requests
	Concurrency int `short:"c" long:"concurrency" default:"25" description:"Concurrency For Requests"`
}

func main() {
	var wg sync.WaitGroup
	var hook_url string
	EnvVars := make(map[string]string)
	_, err := flags.Parse(&Options)
	if err != nil {
		return
	}
	env(EnvVars)
	if err != nil {
		panic(err)
	}
	// get the value of flags and get webhook url from .env file
	data := Options.Data
	conc := Options.Concurrency
	channel := Options.Channel
	if channel == "" {
		hook_url = EnvVars["SLACK_WEBHOOK"]
	} else {
		hook_url = EnvVars["SLACK_WEBHOOK_"+channel]
	}

	// Check if user will send message with --data flag or via 'cat file'
	if data == "" {
		urls := getUrls()
		for i := 0; i < conc; i++ {
			wg.Add(1)
			go func() {
				// Loop on urls channel and start send message function
				defer wg.Done()
				for url := range urls {
					sendMessage(hook_url, url)
				}
			}()
		}
		// wait untill all goroutines finished
		defer wg.Wait()
		return
	}
	// send meassage with --data flag value
	sendMessage(hook_url, data)

}

// Read urls from std input in case user do 'cat file'
func getUrls() <-chan string {
	// create urls channel
	urls := make(chan string)
	// scan all std
	scan := bufio.NewScanner(os.Stdin)
	go func() {
		defer close(urls)
		for scan.Scan() {
			// send every line to urls channel
			urls <- scan.Text()
		}
	}()
	return urls
}

// Read urls from std input in case user do 'cat file'
func env(vars map[string]string) {
	// create env map
	dirname, _ := os.UserHomeDir()
	file, err := os.Open(filepath.Join(dirname, ".env"))
	if err != nil {
		fmt.Println(err)
	}
	// scan all std
	scan := bufio.NewScanner(file)
	defer file.Close()
	for scan.Scan() {
		line := scan.Text()
		// send every line to urls channel
		nw := strings.Split(line, "=")
		vars[nw[0]] = nw[1]
	}
}

func sendMessage(hook_url string, message string) {
	// create the json object for message
	values := map[string]string{"text": message}
	jsonValue, _ := json.Marshal(values)
	// create the request and convert the map into buffers
	req, err := http.NewRequest("POST", hook_url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return
	}
	// Start The Request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
