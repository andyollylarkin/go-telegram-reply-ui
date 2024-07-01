package reply

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var onceCreate sync.Once

var replyMessages map[int]OnReplyCallback

var msgMu sync.RWMutex

type OnReplyCallback bot.HandlerFunc

type Reply struct {
	bot                   *bot.Bot
	allowWithoutReply     bool
	inputFieldPlaceholder string
	parseMode             models.ParseMode
}

func New(bot *bot.Bot) *Reply {
	return &Reply{
		bot:               bot,
		allowWithoutReply: false,
		parseMode:         models.ParseModeHTML,
	}
}

func (r *Reply) AllowSendingWithoutReply() *Reply {
	r.allowWithoutReply = true

	return r
}

func (r *Reply) WithInputFieldPlaceholber(text string) *Reply {
	r.inputFieldPlaceholder = text

	return r
}

func (r *Reply) WithParseMode(mode models.ParseMode) *Reply {
	r.parseMode = mode

	return r
}

func (r *Reply) Send(toChat int64, messageText string, onReply OnReplyCallback) error {
	if onReply == nil {
		return fmt.Errorf("onReply must be set. Nil")
	}
	if r.bot == nil {
		return fmt.Errorf("bot instance must be set. Nil")
	}

	onceCreate.Do(func() {
		replyMessages = make(map[int]OnReplyCallback, 0)
		r.bot.RegisterHandlerMatchFunc(func(update *models.Update) bool {
			if update == nil || update.Message == nil || update.Message.ReplyToMessage == nil {
				return false
			}

			msgMu.RLock()
			defer msgMu.RUnlock()

			_, ok := replyMessages[update.Message.ReplyToMessage.ID]
			if !ok {
				return false
			}

			return true
		}, func(ctx context.Context, botClient *bot.Bot, update *models.Update) {
			if update == nil || update.Message == nil {
				return
			}

			msgMu.Lock()
			defer msgMu.Unlock()

			defer delete(replyMessages, update.Message.ID)

			replyCallback, ok := replyMessages[update.Message.ReplyToMessage.ID]
			if !ok {
				return
			}

			replyCallback(ctx, r.bot, update)

			// delete reply messages
			botClient.DeleteMessage(ctx, &bot.DeleteMessageParams{ // bot send
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ReplyToMessage.ID,
			})
			botClient.DeleteMessage(ctx, &bot.DeleteMessageParams{ // user reply
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.ID,
			})
		})
	})

	m, err := r.bot.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID:    toChat,
		Text:      messageText,
		ParseMode: r.parseMode,
		ReplyMarkup: &models.ForceReply{
			ForceReply:            true,
			InputFieldPlaceholder: r.inputFieldPlaceholder,
		},
		ReplyParameters: &models.ReplyParameters{
			AllowSendingWithoutReply: r.allowWithoutReply,
		},
	})
	if err != nil {
		return err
	}

	msgMu.Lock()
	defer msgMu.Unlock()

	replyMessages[m.ID] = onReply

	return nil
}
