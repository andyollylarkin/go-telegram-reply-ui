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
}

func New(bot *bot.Bot) *Reply {
	return &Reply{
		bot:               bot,
		allowWithoutReply: false,
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
			if update == nil || update.Message == nil {
				return false
			}

			msgMu.RLock()
			defer msgMu.RUnlock()

			_, ok := replyMessages[update.Message.ReplyToMessage.ID]
			if !ok {
				return false
			}

			return true
		}, func(ctx context.Context, bot *bot.Bot, update *models.Update) {
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

			replyCallback(context.Background(), r.bot, update)
		})
	})

	m, err := r.bot.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: toChat,
		Text:   messageText,
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