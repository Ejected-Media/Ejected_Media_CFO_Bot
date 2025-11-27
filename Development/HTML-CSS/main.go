package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
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
	ctx := context.Background()
	opt := option.WithCredentialsFile("serviceAccountKey.json") // Ensure this file exists
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
	token := "YOUR_DISCORD_TOKEN_HERE" // TODO: Move to Env Var
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("✅ CFO Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// COMMAND: !spend <amount> <category> #Project
	if strings.HasPrefix(m.Content, "!spend") {
		args := strings.Fields(m.Content)
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "❌ Usage: `!spend <amount> <category> #Project`")
			return
		}

		amountStr := args[1]
		category := args[2]
		project := "General"

		// Extract Project Tag
		for _, word := range args {
			if strings.HasPrefix(word, "#") {
				project = strings.TrimPrefix(word, "#")
			}
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Invalid amount.")
			return
		}

		err = recordTransaction(m.Author.Username, amount, category, project)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Database error.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ **Saved:** $%.2f for *%s* [Project: **%s**]", amount, category, project))
	}

	// COMMAND: !balance
	if strings.TrimSpace(m.Content) == "!balance" {
		ctx := context.Background()
		iter := fsClient.Collection("transactions").Documents(ctx)
		docs, err := iter.GetAll()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "❌ Error retrieving data.")
			return
		}

		var total float64 = 0
		var count int = 0

		for _, doc := range docs {
			data := doc.Data()
			if amount, ok := data["amount"].(float64); ok {
				total += amount
				count++
			}
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("💰 **Total Spend:** $%.2f (%d txns)", total, count))
	}
}

func recordTransaction(user string, amount float64, category string, project string) error {
	ctx := context.Background()
	_, _, err := fsClient.Collection("transactions").Add(ctx, map[string]interface{}{
		"amount":    amount,
		"category":  category,
		"project":   project,
		"user":      user,
		"currency":  "USD",
		"timestamp": firestore.ServerTimestamp,
	})
	return err
}
