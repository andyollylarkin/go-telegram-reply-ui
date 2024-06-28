package main

import (
	"context"
	"log"

	reply "github.com/andyollylarkin/go-telegram-reply-ui"
	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func main() {
	token := "<YOUR_BOT_TOKEN>"
	b, err := tgbot.New(token)
	if err != nil {
		log.Fatal(err)
	}

	b.RegisterHandler(tgbot.HandlerTypeMessageText, "/action", tgbot.MatchTypeExact,
		func(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
			repl := reply.New(b)

			repl.WithInputFieldPlaceholber("Search").Send(update.Message.Chat.ID, "Search 🔎", onReply)
		})

	b.Start(context.Background())
}

func onReply(ctx context.Context, bot *tgbot.Bot, update *models.Update) {
	bot.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Вы искали: " + update.Message.Text,
	})
}
