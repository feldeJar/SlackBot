package main

import (
	"strings"
	"fmt"
	"github.com/nlopes/slack"
)
const (
	ChatName = "general"
	token = "xoxb-48846318512-365367542772-qjXOTQ0pHbUVtcxllzfs8g1C"

)

func main() {
	for{
		var rtm,msg,prefix,api = usrInput(token)
		respond(rtm,msg,prefix,api)
	}

}
//TODO: place this in a method that returns txt message and timestamp
	func usrInput(token string) (rtm*slack.RTM, msg*slack.MessageEvent, prefix string, api * slack.Client) {
		api = slack.New(token)
		api.SetDebug(true)
		rtm = api.NewRTM()
		go rtm.ManageConnection()
		for {
			select {
			case msg := <-rtm.IncomingEvents:
				fmt.Print("Event Received: ")

				switch ev := msg.Data.(type) {
				case *slack.ConnectedEvent:
					fmt.Println("Connection counter:", ev.ConnectionCount)
				case *slack.MessageEvent:
					fmt.Printf("Message: %v\n", ev)
					info := rtm.GetInfo()
					prefix := fmt.Sprintf("<@%s> ", info.User.ID)
					if ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) {
						return rtm,ev,prefix,api
					}

				case *slack.RTMError:
					fmt.Printf("Error: %s\n", ev.Error())
				case *slack.InvalidAuthEvent:
					fmt.Printf("Invalid credentials")
					break
				default:

				}
			}
		}
	}
	//respond to posted question, no return value
	func respond(rtm*slack.RTM, msg*slack.MessageEvent, prefix string, api * slack.Client) {
	var response string
	ts := msg.Timestamp
	text := msg.Text
	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	acceptedGreetings := map[string]bool{
	"what's up?": true,
	"hey!":       true,
	"yo":         true,
	}
	acceptedHowAreYou := map[string]bool{
	"how's it going?": true,
	"how are ya?":     true,
	"feeling okay?":   true,
	}
	acceptedPing := map[string]bool{
	"ping": true,
	"Ping": true,
	}
	if acceptedPing[text]{
	response = "Pong"
	api.PostMessage(msg.Channel, response, slack.PostMessageParameters{ThreadTimestamp:ts})
	}
	if acceptedGreetings[text] {
	response = "What's up buddy!?!?!"
	api.PostMessage(msg.Channel, response, slack.PostMessageParameters{ThreadTimestamp:ts})
	} else if acceptedHowAreYou[text] {
	response = "Good. How are you?"
	api.PostMessage(msg.Channel, response, slack.PostMessageParameters{ThreadTimestamp:ts})
	}
	}

