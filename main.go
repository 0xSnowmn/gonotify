package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	flags "github.com/jessevdk/go-flags"
	env "github.com/joho/godotenv"
)

// Setting Options
var Options struct {
	// File For urls
	Data string `short:"d" long:"data" description:"data sends to channel" required:"false"`
	// Concurrency For the requests
	Concurrency int `short:"c" long:"concurrency" default:"25" description:"Concurrency For Requests"`
}

func main() {
	var wg sync.WaitGroup
	_, err := flags.Parse(&Options)
	if err != nil {
		return
	}
	err = env.Load()
	if err != nil {
		fmt.Println("Error in Loading Env File")
		return
	}
	// get the value of flags and get webhook url from .env file
	data := Options.Data
	conc := Options.Concurrency
	hook_url := os.Getenv("SLACK_WEBHOOK")

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
