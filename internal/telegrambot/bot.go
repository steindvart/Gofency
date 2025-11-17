package telegrambot

import (
	"context"
	"log"

	"gofency/internal/captcha"
	"gofency/internal/fsm"
	"gofency/internal/localization"
	"gofency/internal/repositories"
	"gofency/internal/telegrambot/handlers"
	"gofency/internal/telegrambot/middlewares"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Bot struct {
	api            *bot.Bot
	localization   *localization.Service
	userRepository repositories.UserRepository
	captchaService *captcha.Service
	captchaFSM     *fsm.CaptchaFSM
}

type Config struct {
	Token               string
	LocalizationService *localization.Service
	UserRepository      repositories.UserRepository
	CaptchaService      *captcha.Service
	CaptchaFSM          *fsm.CaptchaFSM
}

func NewBot(cfg Config) (*Bot, error) {
	localizationMiddleware := middlewares.NewLocalization(cfg.LocalizationService, cfg.UserRepository)

	userRepositoryMiddleware := func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			ctx = repositories.WithUserRepository(ctx, cfg.UserRepository)
			next(ctx, b, update)
		}
	}

	captchaFSMMiddleware := func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			ctx = fsm.WithCaptchaFSM(ctx, cfg.CaptchaFSM)
			next(ctx, b, update)
		}
	}

	opts := []bot.Option{
		bot.WithMiddlewares(
			userRepositoryMiddleware,
			captchaFSMMiddleware,
			localizationMiddleware.Handler,
			middlewares.LogMessageWithText,
		),

		// bot.WithMessageTextHandler("start", bot.MatchTypeCommand, handlers.CommandStart),
		// bot.WithMessageTextHandler("help", bot.MatchTypeCommand, handlers.CommandHelp),
		// bot.WithMessageTextHandler("lang", bot.MatchTypeCommand, handlers.CommandLanguage),
		// bot.WithMessageTextHandler("testcaptcha", bot.MatchTypeCommand, handlers.CommandTestCaptcha(cfg.CaptchaService)),

		// bot.WithCallbackQueryDataHandler("set_lang_", bot.MatchTypePrefix, handlers.HandleLanguageCallback),

		// Handle new chat members and text messages for captcha verification
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			// Log the update for debugging
			if update.Message != nil {
				log.Printf("Default handler received message from user %d in chat %d",
					update.Message.From.ID, update.Message.Chat.ID)
			}

			if update.Message != nil && update.Message.NewChatMembers != nil {
				log.Printf("New chat members detected: %d members", len(update.Message.NewChatMembers))
				handlers.HandleNewChatMember(cfg.CaptchaService)(ctx, b, update)
				return
			}
			// Check if this is a text message that might be a captcha answer
			// Skip if it's a command (starts with /)
			if update.Message != nil && update.Message.Text != "" && len(update.Message.Text) > 0 && update.Message.Text[0] != '/' {
				handlers.HandleCaptchaTextAnswer(ctx, b, update)
				return
			}
		}),
	}

	b, err := bot.New(cfg.Token, opts...)
	if err != nil {
		return nil, err
	}

	return &Bot{
		api:            b,
		localization:   cfg.LocalizationService,
		userRepository: cfg.UserRepository,
		captchaService: cfg.CaptchaService,
		captchaFSM:     cfg.CaptchaFSM,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	log.Println("Starting Telegram bot...")
	log.Printf("Supported languages: %v", b.localization.SupportedLanguages())

	b.api.Start(ctx)

	return nil
}
