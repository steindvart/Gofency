package handlers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"gofency/internal/captcha"
	"gofency/internal/fsm"
	"gofency/internal/localization"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
)

func HandleNewChatMember(captchaService *captcha.Service) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		log.Printf("HandleNewChatMember called")

		if update.Message == nil || len(update.Message.NewChatMembers) == 0 {
			log.Printf("No message or no new chat members")
			return
		}

		chatID := update.Message.Chat.ID
		log.Printf("Processing %d new member(s) in chat %d", len(update.Message.NewChatMembers), chatID)

		for _, newMember := range update.Message.NewChatMembers {
			if newMember.IsBot {
				log.Printf("Skipping bot: %s", newMember.Username)
				continue
			}

			log.Printf("Processing new member: %s (ID: %d)", newMember.FirstName, newMember.ID)

			// Generate captcha
			captchaImg, err := captchaService.Generate()
			if err != nil {
				log.Printf("Failed to generate captcha: %v", err)
				continue
			}

			log.Printf("Generated captcha with answer: %s", captchaImg.Answer)

			username := GenerateMention(update.Message.From)

			// Send welcome message with captcha
			welcomeText := localization.GetText(ctx, "captcha_welcome", map[string]any{
				"Username": username,
			})
			welcomeText += "\n\n" + localization.GetText(ctx, "captcha_prompt", nil)

			log.Printf("Sending captcha image to chat %d", chatID)

			// Send captcha image
			photoMsg, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
				ChatID:    chatID,
				Photo:     &tgmodels.InputFileUpload{Data: bytes.NewReader(captchaImg.Image)},
				Caption:   welcomeText,
				ParseMode: tgmodels.ParseModeMarkdownV1,
			})
			if err != nil {
				log.Printf("Failed to send captcha image: %v", err)
				continue
			}

			log.Printf("Captcha image sent, message ID: %d", photoMsg.ID)

			// Get FSM from context
			captchaFSM, ok := fsm.GetCaptchaFSM(ctx)
			if !ok {
				log.Printf("Captcha FSM not found in context")
				continue
			}

			// Save state in FSM
			captchaFSM.SetState(newMember.ID, &fsm.CaptchaData{
				ChatID:         chatID,
				UserID:         newMember.ID,
				Answer:         captchaImg.Answer,
				ExpiresAt:      time.Now().Add(30 * time.Second),
				PhotoMessageID: photoMsg.ID,
			})

			log.Printf("FSM state saved for user %d", newMember.ID)

			// Schedule timeout check
			go scheduleTimeoutCheck(b, chatID, newMember.ID, username, photoMsg.ID, captchaFSM)

			log.Printf("Timeout check scheduled for user %d", newMember.ID)
		}
	}
}

// scheduleTimeoutCheck checks if user completed captcha within timeout
func scheduleTimeoutCheck(b *bot.Bot, chatID, userID int64, username string, photoMsgID int, captchaFSM *fsm.CaptchaFSM) {
	time.Sleep(30 * time.Second)

	// Check if state still exists (if it does, user didn't complete it)
	if _, ok := captchaFSM.GetState(userID); !ok {
		// User already verified or removed
		return
	}

	// Check if expired
	if !captchaFSM.IsExpired(userID) {
		// Still within time window
		return
	}

	ctx := context.Background()

	// Delete state
	captchaFSM.DeleteState(userID)

	// Kick and ban user
	_, err := b.BanChatMember(ctx, &bot.BanChatMemberParams{
		ChatID:         chatID,
		UserID:         userID,
		UntilDate:      int(time.Now().Add(10 * time.Minute).Unix()),
		RevokeMessages: false,
	})
	if err != nil {
		log.Printf("Failed to ban user %d: %v", userID, err)
		return
	}

	// Delete captcha messages
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: photoMsgID,
	})

	// Send timeout message
	timeoutText := fmt.Sprintf("⏱ Verification timeout. %s has been removed from the chat and banned for 10 minutes.", username)

	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      timeoutText,
		ParseMode: tgmodels.ParseModeMarkdownV1,
	})
	if err != nil {
		log.Printf("Failed to send timeout message: %v", err)
		return
	}

	// Delete timeout message after 10 seconds
	time.Sleep(10 * time.Second)
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: msg.ID,
	})
}
