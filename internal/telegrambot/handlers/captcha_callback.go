package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gofency/internal/fsm"
	"gofency/internal/localization"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

// HandleCaptchaTextAnswer handles captcha text input from users
func HandleCaptchaTextAnswer(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	answer := update.Message.Text

	// Get FSM from context
	captchaFSM, ok := fsm.GetCaptchaFSM(ctx)
	if !ok {
		// No FSM available, ignore message
		return
	}

	// Check if user has a pending captcha
	data, ok := captchaFSM.GetState(userID)
	if !ok {
		// No pending captcha for this user, ignore message
		return
	}

	// Check if expired
	if captchaFSM.IsExpired(userID) {
		// Already expired, will be handled by timeout goroutine
		return
	}

	// Delete the user's answer message
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: update.Message.ID,
	})

	// Validate answer
	if answer == data.Answer {
		// Correct answer - delete state
		captchaFSM.DeleteState(userID)

		// Delete captcha messages
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: data.PhotoMessageID,
		})

		// Send success message
		successText := localization.GetText(ctx, "captcha_success", map[string]interface{}{
			"Username": GenerateMention(update.Message.From),
		})

		msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      successText,
			ParseMode: tgmodels.ParseModeMarkdownV1,
		})
		if err != nil {
			log.Printf("Failed to send success message: %v", err)
			return
		}

		// Delete success message after 10 seconds
		go func() {
			time.Sleep(10 * time.Second)
			b.DeleteMessage(context.Background(), &bot.DeleteMessageParams{
				ChatID:    chatID,
				MessageID: msg.ID,
			})
		}()
	} else {
		// Wrong answer - delete state
		captchaFSM.DeleteState(userID)

		// Kick user (ban then immediately unban to just kick)
		_, err := b.BanChatMember(ctx, &bot.BanChatMemberParams{
			ChatID:         chatID,
			UserID:         userID,
			UntilDate:      int(time.Now().Add(10 * time.Minute).Unix()),
			RevokeMessages: false,
		})
		if err != nil {
			log.Printf("Failed to ban user %d: %v", userID, err)
		}

		// Delete captcha messages
		b.DeleteMessage(ctx, &bot.DeleteMessageParams{
			ChatID:    chatID,
			MessageID: data.PhotoMessageID,
		})

		// Send failure message
		failedText := localization.GetSimpleText(ctx, "captcha_failed")
		msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   failedText,
		})
		if err != nil {
			log.Printf("Failed to send failure message: %v", err)
			return
		}

		// Delete failure message after 10 seconds
		go func() {
			time.Sleep(10 * time.Second)
			b.DeleteMessage(context.Background(), &bot.DeleteMessageParams{
				ChatID:    chatID,
				MessageID: msg.ID,
			})
		}()
	}
}

func EscapeMarkdown(text string) string {
	specialChars := "_*[]()~`>#+-=|{}.!"
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, string(char), "\\"+string(char))
	}
	return text
}

func GenerateMention(user *tgmodels.User) string {
	if user.FirstName != "" && user.LastName != "" {
		return fmt.Sprintf("[%s %s](tg://user?id=%d)", EscapeMarkdown(user.FirstName), EscapeMarkdown(user.LastName), user.ID)
	} else if user.FirstName != "" {
		return fmt.Sprintf("[%s](tg://user?id=%d)", EscapeMarkdown(user.FirstName), user.ID)
	} else if user.Username != "" {
		return fmt.Sprintf("[@%s](tg://user?id=%d)", EscapeMarkdown(user.FirstName), user.ID)
	}
	return fmt.Sprintf("[User](tg://user?id=%d)", user.ID)
}
