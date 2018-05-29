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
	token    = "xoxb-48846318512-365367542772-qjXOTQ0pHbUVtcxllzfs8g1C" //bot token
)

func main() {
	//TODO get better names
	rtm, msg, usrString, api := usrInput(token)
	dTreeVal, dTreeMap := ExtractKeyWords(usrString)
	keyWords, _ := RankKeyWords(dTreeVal, dTreeMap)
	respond(rtm, msg, api, keyWords)
}

//Pulls String from Slack when bot gets called
func usrInput(token string) (rtm *slack.RTM, msg *slack.MessageEvent, inputString string, api *slack.Client) {
	//setup listener
	api = slack.New(token)
	api.SetDebug(true)
	rtm = api.NewRTM()
	go rtm.ManageConnection()
	//loop to wait for userInput
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
					inputString := strings.TrimPrefix(ev.Text, prefix)
					return rtm, ev, inputString, api
				}
				//err evaluation
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

//Post message in slack with same TS, Takes []string as Input
func respond(rtm *slack.RTM, msg *slack.MessageEvent, api *slack.Client, response []string) {
	ts := msg.Timestamp
	//place for loop to respond to more than one call
	api.PostMessage(msg.Channel, strings.Join(response, "|"), slack.PostMessageParameters{ThreadTimestamp: ts})
}

//used to classify a string > 20 length
func ClassifyText(usrtext string) {
	ctx := context.Background()
	// Creates a client.
	client, err := language.NewClient(ctx)
	if err != nil {
		fmt.Println("error:", err)
		log.Fatalf("Failed to create client: %v", err)
	}
	// Sets the text to analyze.
	text := usrtext
	classText, err := client.ClassifyText(ctx, &languagepb.ClassifyTextRequest{
		Document: &languagepb.Document{
			Type: languagepb.Document_PLAIN_TEXT,
			Source: &languagepb.Document_Content{
				Content: text,
			},
		},
		EncodingType: languagepb.EncodingType_UTF8,
	}, )
	if err != nil {
		fmt.Println("error:", err)
		log.Fatalf("Failed to analyze text: %v", err)
	}
	for _, classText := range classText.GetCategories() {
		fmt.Println("inside ran")
		fmt.Println("classText.Name=", classText.Name)
	}
	client.Close()
}

// GoogleCloud MetaData, Currenlty set up to return Wiki URL link
func AnalyzeEntitiesMetadata(usrtext string) (entity string) {
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
		fmt.Println("Entities:", entity)
		return (entity.Metadata["wikipedia_url"])
	}
	client.Close()
	return ("No Wikipedia Page Found")

}

// Takes input and evaluates what the key words are and rank of key words
func ExtractKeyWords(text string) ([]int32, [] string) {
	//set up client
	ctx := context.Background()
	client, err := language.NewClient(ctx)
	//err report
	if err != nil {
		fmt.Println("error:", err)
		log.Fatalf("Failed to create client: %v", err)
	}
	//GoogleCloud Keyword evaluation may need tweaking
	entity, err := client.AnnotateText(ctx, &languagepb.AnnotateTextRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: text,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		Features: &languagepb.AnnotateTextRequest_Features{
			ExtractSyntax: true,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	})

	TreeKey := make([]int32, len(entity.Tokens)) //dynamic [] lengths
	TreeMap := make([]string, len(entity.Tokens))
	// pulls out the DTree value along with the corresponding word
	for i := 0; i < len(entity.Tokens); i++ {
		//OUTPUT TESTERS
		//fmt.Println("-->",entity.Tokens[i].DependencyEdge.HeadTokenIndex)
		//fmt.Print("-->",entity.Tokens[entity.Tokens[i].DependencyEdge.HeadTokenIndex].Text.Content)
		TreeKey[i] = entity.Tokens[i].DependencyEdge.HeadTokenIndex
		TreeMap[i] = entity.Tokens[i].Text.Content

	}
	return TreeKey, TreeMap
}

//Ranks string with the google language cloud, retuns key words and rank(deffined by # of DTree Links
func RankKeyWords(intKey []int32, strMap []string) (key []string, rank []int) {
	temp := make([]int, len(strMap)) //set rank array length to max
	for i := 0; i < len(temp); i++ { //fill temp with input from user
		temp[intKey[i]] += 1
	}
	getKeys := make([]string, 0) //dynamic string[]
	getRank := make([]int, 0)    //dynamic int[]

	for i := len(temp) - 1; i >= 0; i-- { //runs though array and puts the D-Tree ranks in order
		for j := 0; j <= len(temp)-1; j++ {
			if (temp[j] == i && temp[j] != 0) {
				//fmt.Println("Location",j)TESTER
				getKeys = append(getKeys, strMap[j]) //places the keywords in order of rank
				//fmt.Println("Value",temp[j])TESTER
				getRank = append(getRank, temp[j])
			}
		}
	}
	//output rank evaluation
	fmt.Println("Rank of tree dependencies:",getRank)
	return getKeys, getRank

}
