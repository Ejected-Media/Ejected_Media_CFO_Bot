This is the foundational step. We are creating the "spinal cord" of the bot: a Go program that listens to Discord events and has a secure line to your Firebase database.
Prerequisites
Before running this code, you will need to initialize a Go module and get your credentials ready.
 * Discord Token: From the Discord Developer Portal.
 * Firebase Credentials: A serviceAccountKey.json file from your Firebase Project Settings > Service Accounts.
The "Walking Skeleton" (main.go)
This code initializes the connection to both services. If it works, the bot will come online and print "CFO Bot is ready" to your console.

```
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/bwmarrin/discordgo"
	"google.golang.org/api/option"
)

// Global Firestore Client
var fsClient *firestore.Client

func main() {
	// 1. Initialize Firebase
	// We use a context to manage the connection lifecycle
	ctx := context.Background()
	
	// Ensure you have 'serviceAccountKey.json' in your root directory
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error initializing firestore client: %v", err)
	}
	fsClient = client
	defer fsClient.Close()
	fmt.Println("✅ Firebase Firestore Connected")

	// 2. Initialize Discord
	// Replace with your actual token or load from os.Getenv("DISCORD_TOKEN")
	token := "YOUR_DISCORD_BOT_TOKEN_HERE" 
	
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("✅ CFO Bot is now running. Press CTRL-C to exit.")

	// Wait here until CTRL-C or other term signal is received.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Simple "Ping" command to test the loop
	if strings.TrimSpace(m.Content) == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong! 🏓 (System Online)")
		
		// TEST: Log this interaction to Firebase to prove the link works
		// We run this in a goroutine so we don't block the chat
		go logInteraction(m.Author.Username, m.Content)
	}
}

func logInteraction(username string, message string) {
	ctx := context.Background()
	_, _, err := fsClient.Collection("logs").Add(ctx, map[string]interface{}{
		"user":    username,
		"message": message,
		"timestamp": firestore.ServerTimestamp,
	})
	if err != nil {
		log.Printf("Failed to log to Firebase: %v", err)
	} else {
		log.Printf("Logged interaction from %s to Firebase", username)
	}
}
```

Setup Instructions
To get this running locally:
 * Initialize the Module:
   go mod init cfo-bot

 * Get Dependencies:
   go get github.com/bwmarrin/discordgo
go get firebase.google.com/go
go get google.golang.org/api/option

 * Run It:
   go run main.go

What this achieves
When you type !ping in your Discord server:
 * Discord sends the event to your Go app.
 * Go replies "Pong!".
 * Crucially: Go silently writes a document to your logs collection in Firestore.
This proves that Discord -> Go -> Firestore is fully operational.
Next Step: Once you confirm you can get a "Pong", would you like to move on to building the "Expense Injection" command, where we parse a string like !spend 50 ServerCosts and save it to the actual ledger?
