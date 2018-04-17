package main

import (
	"encoding/json"
	"fmt"
	"homecast"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func main() {
	app := cli.NewApp()
	app.Name = "google calendar notifier"
	app.Usage = "notifier google calendar event with google home."
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name: "calendar, c",
			// Value: &cli.StringSlice{"primary"},
			Usage: "set your calendar id.",
		},
		cli.Int64Flag{
			Name:  "interval, i",
			Value: 5,
			Usage: "set notifier interval.",
		},
	}
	app.Action = func(c *cli.Context) error {
		calendars := c.StringSlice("calendar")
		if len(calendars) == 0 {
			calendars = []string{"primary"}
		}
		runNotifierAgent(calendars, c.Int64("interval"))
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// 前回の通知実行時間.
var timepre = time.Now().Add(time.Minute * -5)

func runNotifierAgent(calendarIDs []string, interval int64) {
	// TODO go-scheduler
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	srv, err := calendar.New(getClient(config))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	timemin := timepre
	timemax := time.Now().Add(1 * time.Minute)

	var eventitems []*calendar.Event
	for _, calendarID := range calendarIDs {
		events, err := srv.Events.List(calendarID).ShowDeleted(false).
			SingleEvents(true).TimeMin(timemin.Format(time.RFC3339)).
			TimeMax(timemax.Format(time.RFC3339)).OrderBy("startTime").Do()
		if err != nil {
			fmt.Printf("Unable to retrieve next ten of the user's events: %v", err)
		} else {
			eventitems = append(eventitems, events.Items...)
		}
	}

	var summarys []string
	for _, item := range eventitems {
		date, err := time.Parse(time.RFC3339, item.Start.DateTime)
		if err != nil {
			fmt.Printf("Unable to parse Start.DateTime: %v", item.Start.DateTime)
			continue
		}
		// TODO 全日処理
		// if date == "" {
		// 	date = item.Start.Date
		// }

		if date.Before(timemax) && date.After(timemin) {
			// TODO 文字数制限
			summarys = append(summarys, item.Summary)
		}
	}

	if len(summarys) == 0 {
		fmt.Println("No upcoming events found.")
		return
	}

	ctx := context.Background()
	devices := homecast.LookupAndConnect(ctx)

	notifiertext := strings.Join(summarys, ",")
	for _, device := range devices {
		if err := device.Speak(ctx, fmt.Sprintf("%vの時間です\n", notifiertext), "ja"); err != nil {
			fmt.Printf("Failed to speak: %v", err)
		}
	}
	timepre = time.Now()
}
