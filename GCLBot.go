package main
// Language Bot
import (
	"log"
	// Imports the Google Cloud Natural Language API client package.
	"cloud.google.com/go/language/apiv1"
	"golang.org/x/net/context"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
	"fmt"
	"github.com/nlopes/slack"
	"strings"
)
const (
	ChatName = "general"
	token = "xoxb-48846318512-365367542772-qjXOTQ0pHbUVtcxllzfs8g1C"
)
func main() {
		rtm, msg, message, api := usrInput(token)
		reply := AnalyzeEntitiesMetadata(message)
		respond(rtm, msg, api, reply)
}
func usrInput(token string) (rtm*slack.RTM, msg*slack.MessageEvent, inputString string, api * slack.Client) {
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
				prefix := fmt.Sprintf("<@%s>", info.User.ID)
				if ev.User != info.User.ID && strings.HasPrefix(ev.Text, prefix) {
					inputString:= strings.TrimPrefix(ev.Text, prefix)
					return rtm,ev,inputString,api
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
func respond(rtm*slack.RTM, msg*slack.MessageEvent,api * slack.Client,response string) {
	ts := msg.Timestamp
	api.PostMessage(msg.Channel, response, slack.PostMessageParameters{ThreadTimestamp: ts})
}
func ClassifyText (usrtext string){
	ctx := context.Background()
	// Creates a client.
	client, err := language.NewClient(ctx)
	if err != nil {
		fmt.Println("error:",err)
		log.Fatalf("Failed to create client: %v", err)
	}
	// Sets the text to analyze.
	text := usrtext

	classText, err := client.ClassifyText(ctx,&languagepb.ClassifyTextRequest{
		Document:&languagepb.Document{
			Type: languagepb.Document_PLAIN_TEXT,
			Source: &languagepb.Document_Content{
				Content: text,
			},
		},
		EncodingType:languagepb.EncodingType_UTF8,
	},)
	if err != nil {
		fmt.Println("error:",err)
		log.Fatalf("Failed to analyze text: %v", err)
	}
	for _, classText := range classText.GetCategories(){
		fmt.Println("inside ran")
		fmt.Println("classText.Name=",classText.Name)
	}
	client.Close()
}
func AnalyzeEntitiesMetadata(usrtext string)(entity string){
	ctx := context.Background()
	// Creates a client.
	client, err := language.NewClient(ctx)
	if err != nil {
		fmt.Println("Failed to create client:", err)
		log.Fatalf("Failed to create client: %v", err)
	}
	// Sets the text to analyze.
	text := usrtext
	// Detects the sentiment of the text.
	entities, err := client.AnalyzeEntities(ctx, &languagepb.AnalyzeEntitiesRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: text,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	})


	if err != nil {
		fmt.Println("Failed to analyze text: %v", err)
		log.Fatalf("Failed to analyze text: %v", err)
	}
	for _, entity := range entities.GetEntities() {
		/*fmt.Println("entity.Name=",entity.Name)
		fmt.Println("entity.Type=",entity.Type)
		fmt.Println("entity.Mentions=",entity.Mentions)
		fmt.Println("URL:",entity.Metadata["wikipedia_url"])
		//url:= make(map[string])
			//fmt.Println("URL:",url)
		//fmt.Println("____________")*/
		return (entity.Metadata["wikipedia_url"])
	}
	client.Close()
	return ("No Wikipedia Page Found")

}
