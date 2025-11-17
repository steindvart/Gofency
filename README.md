# 🛡️ Gofency Bot
Telegram bot for protecting group chats from spam.

> Made with ❤️ to keep Telegram groups safe and clean.

## 🎯 What It Does
Gofency Bot helps keep your Telegram group chats clean and safe by:

- **CAPTCHA Verification**: New users must solve a CAPTCHA before they can participate in the chat
- **Automatic Moderation**: Users who fail CAPTCHA verification are automatically removed from the group
- **Multilingual Support**: Built-in support for multiple languages (Russian, English, and more)
- **Spam Prevention**: Protects your community from automated bot attacks

## ✨ Key Features
- 🔐 **Image-based CAPTCHA** - Visual verification to prevent automated bot attacks
- 🌍 **Multi-language Support** - Automatically adapts to user's language preferences

## 🚀 Quick Start
1. Add the `@gofency_bot` to your Telegram group.
2. Set administrator rights for the bot.
3. Ready! Gofency protects your group!

## 📋 Planned Features
We're constantly working to improve Gofency Bot. Here's what's coming next:

### 🔧 Customization Options
- [ ] **Join/Leave Message Management** - Option to automatically delete user join/leave system messages
- [ ] **Configurable Timeouts** - Customize CAPTCHA response time, ban duration, and ban policies
- [ ] **Restriction Mode** - Instead of immediate kick, restrict user rights (read-only) to allow them to understand why they were flagged

### 🎨 CAPTCHA Improvements
- [ ] **Enhanced CAPTCHA Generation** - More diverse and complex CAPTCHA images
- [ ] **Multiple CAPTCHA Types** - Support for different verification methods (math problems, image selection, etc.)

### 🛡️ Advanced Protection
- [ ] **Spam Detection** - AI-powered message analysis to detect spam and scam content
- [ ] **Flexible Action Policies** - Configurable actions for different types of violations
- [ ] **Content Filtering** - Detect and handle:

## 🛠️ Technology Stack
- **Language**: Go 1.25
- **Database**: PostgreSQL 15
- **Bot Framework**: [go-telegram/bot](https://github.com/go-telegram/bot)
- **ORM**: GORM
- **Deployment**: Docker & Docker Compose

## ⭐ Give your Star!
If you find this project useful, please consider giving it a star!
