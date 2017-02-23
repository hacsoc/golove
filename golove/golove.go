/*
This is a command-line application for sending love.
*/
package main

import (
	"fmt"
	"github.com/hacsoc/golove/love"
	"os"
	"strings"
)

func main() {
	api_key := os.Getenv("LOVE_API_KEY")
	base_url := os.Getenv("LOVE_BASE_URL")
	sender := os.Getenv("LOVE_SENDER")
	fmt.Println(api_key)
	if len(os.Args) < 3 {
		fmt.Println("usage: golove recipient[,recipient] message")
		return
	}
	recipient := os.Args[1]
	message := strings.Join(os.Args[2:], " ")
	client := love.NewClient(api_key, base_url)
	err := client.SendLove(sender, recipient, message)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Love sent to %s!", recipient)
	}
}
