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
	"gofency/internal/repositories"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func CommandStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	welcomeText := localization.GetSimpleText(ctx, "welcome_message")

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   welcomeText,
	})
}

func CommandHelp(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	helpText := localization.GetSimpleText(ctx, "help_message")

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   helpText,
	})
}

func CommandLanguage(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	keyboard := models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "🇷🇺 Русский",
					CallbackData: "set_lang_ru",
				},
				{
					Text:         "🇺🇸 English",
					CallbackData: "set_lang_en",
				},
			},
		},
	}

	selectText := localization.GetSimpleText(ctx, "language_selection")

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        selectText,
		ReplyMarkup: keyboard,
	})
}

func HandleLanguageCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.CallbackQuery == nil {
		return
	}

	var langCode, langName string

	switch update.CallbackQuery.Data {
	case "set_lang_ru":
		langCode = "ru"
		langName = "русский"
	case "set_lang_en":
		langCode = "en"
		langName = "English"
	default:
		return
	}

	userRepo, ok := repositories.GetUserRepository(ctx)
	if ok {
		telegramID := update.CallbackQuery.From.ID
		_, err := userRepo.UpsertLanguage(ctx, telegramID, langCode)
		if err != nil {
			log.Printf("Failed to save language preference for user %d: %v\n", telegramID, err)
		}
	}

	confirmText := localization.GetText(ctx, "language_changed", map[string]interface{}{
		"Language": langName,
	})

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            confirmText,
		ShowAlert:       false,
	})

	b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
		Text:      confirmText,
	})
}

// CommandTestCaptcha sends a test captcha to verify the system is working
func CommandTestCaptcha(captchaService *captcha.Service) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		chatID := update.Message.Chat.ID
		userID := update.Message.From.ID

		log.Printf("Test captcha command from user %d in chat %d", userID, chatID)

		// Generate captcha
		captchaImg, err := captchaService.Generate()
		if err != nil {
			log.Printf("Failed to generate captcha: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   fmt.Sprintf("❌ Failed to generate captcha: %v", err),
			})
			return
		}

		log.Printf("Generated test captcha with answer: %s", captchaImg.Answer)

		// Send test message first
		testMsg := fmt.Sprintf("🧪 **Test Captcha Mode**\n\nAnswer: `%s`\n\nThe captcha image will be sent next. Try typing the answer to test the verification system!", captchaImg.Answer)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      testMsg,
			ParseMode: "Markdown",
		})

		// Get username
		username := update.Message.From.FirstName
		if username == "" {
			username = update.Message.From.Username
		}
		if username == "" {
			username = "User"
		}

		// Send captcha image
		welcomeText := fmt.Sprintf("Welcome, %s! Please enter the captcha digits within 30 seconds.", username)

		photoMsg, err := b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID:  chatID,
			Photo:   &models.InputFileUpload{Data: bytes.NewReader(captchaImg.Image)},
			Caption: welcomeText,
		})
		if err != nil {
			log.Printf("Failed to send captcha image: %v", err)
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   fmt.Sprintf("❌ Failed to send captcha image: %v", err),
			})
			return
		}

		// Send prompt message
		promptText := "Please type the digits you see in the captcha image:"

		promptMsg, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   promptText,
		})
		if err != nil {
			log.Printf("Failed to send captcha prompt: %v", err)
			return
		}

		// Get FSM from context
		captchaFSM, ok := fsm.GetCaptchaFSM(ctx)
		if !ok {
			log.Printf("Captcha FSM not found in context")
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: chatID,
				Text:   "❌ FSM not available in context",
			})
			return
		}

		// Save state in FSM
		captchaFSM.SetState(userID, &fsm.CaptchaData{
			ChatID:         chatID,
			UserID:         userID,
			Answer:         captchaImg.Answer,
			ExpiresAt:      time.Now().Add(30 * time.Second),
			PhotoMessageID: photoMsg.ID,
		})

		log.Printf("Test captcha state saved for user %d", userID)

		// Schedule timeout check
		go scheduleTestCaptchaTimeout(b, chatID, userID, username, photoMsg.ID, promptMsg.ID, captchaFSM)
	}
}

// scheduleTestCaptchaTimeout is similar to scheduleTimeoutCheck but for test mode
func scheduleTestCaptchaTimeout(b *bot.Bot, chatID, userID int64, username string, photoMsgID, promptMsgID int, captchaFSM *fsm.CaptchaFSM) {
	time.Sleep(30 * time.Second)

	// Check if state still exists
	if _, ok := captchaFSM.GetState(userID); !ok {
		// User already verified
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

	// Don't ban in test mode, just notify
	// Delete captcha messages
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: photoMsgID,
	})
	b.DeleteMessage(ctx, &bot.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: promptMsgID,
	})

	// Send timeout message
	timeoutText := fmt.Sprintf("⏱ Test timeout! In production mode, %s would be banned for 10 minutes.", username)

	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   timeoutText,
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
