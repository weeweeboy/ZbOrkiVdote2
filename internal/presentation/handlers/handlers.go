package handlers

import (
	"ZbOrkiVdote/internal/application/user/interfaces/usecase"
	"context"
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot     *tgbotapi.BotAPI
	usecase usecase.UserUsecase
	logger  *slog.Logger
}

func NewHandlers(bot *tgbotapi.BotAPI, usecase usecase.UserUsecase, logger *slog.Logger) *Handler {
	return &Handler{
		bot:     bot,
		usecase: usecase,
		logger:  logger,
	}
}

func (h *Handler) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {

		if update.CallbackQuery != nil {
			h.handleCallback(ctx, update)
		} else if update.Message != nil {
			if update.Message.IsCommand() {
				h.handleStartCommand(ctx, update)
			} else {
				h.handleMessage(ctx, update)
			}
		}
	}
}

func (h *Handler) handleStartCommand(ctx context.Context, update tgbotapi.Update) {
	chat := update.Message.Chat
	msg := tgbotapi.NewMessage(chat.ID, "Приветствую тебя, друг❤️! Чтобы найти сборку на нужного тебе героя, просто напиши его имя в чат!😇")
	h.bot.Send(msg)

	msg = tgbotapi.NewMessage(chat.ID, "Для того, чтобы бот работал, нужно ввести полное и правильное имя героя на английском языке")
	h.bot.Send(msg)

	msg = tgbotapi.NewMessage(chat.ID, "Например, Earth Spirit, но есть и коварные имена: Anti-Mage, Nature's Prophet, будь осторожен, дружище!🙂")
	h.bot.Send(msg)

}

func (h *Handler) handleMessage(ctx context.Context, update tgbotapi.Update) {
	chat := update.Message.Chat
	name := update.Message.Text

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("1", fmt.Sprintf("%s:1", name)),
			tgbotapi.NewInlineKeyboardButtonData("2", fmt.Sprintf("%s:2", name)),
			tgbotapi.NewInlineKeyboardButtonData("3", fmt.Sprintf("%s:3", name)),
			tgbotapi.NewInlineKeyboardButtonData("4", fmt.Sprintf("%s:4", name)),
			tgbotapi.NewInlineKeyboardButtonData("5", fmt.Sprintf("%s:5", name)),
		))

	msg := tgbotapi.NewMessage(chat.ID, "Выбери позицию:")
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

func (h *Handler) handleCallback(ctx context.Context, update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		return
	}

	data := update.CallbackQuery.Data

	callbackConfig := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := h.bot.Request(callbackConfig); err != nil {
		h.logger.Error("Не удалось подтверить callback", slog.Any("error", err))
	}

	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		h.logger.Error("Некорректный формат callback data", slog.String("data", data))
		return
	}

	nameOfHero := parts[0]
	positionOfHero := parts[1]

	sborka, err := h.usecase.GetSborka(ctx, nameOfHero, positionOfHero)
	if err != nil {
		h.logger.Error("Ошибка в вызове GetSborka", slog.Any("error", err))

		text := "Ошибка при поиске сборки."
		if strings.Contains(err.Error(), "нет данных") {
			text = "Для этого героя на данной позиции нет данных."
		}

		h.bot.Send(tgbotapi.NewMessage(
			update.CallbackQuery.Message.Chat.ID,
			text,
		))
		return
	}

	for _, part := range splitMessage(sborka, 4000) {
		h.bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, part))
	}
}

func splitMessage(text string, limit int) []string {
	var parts []string

	for len(text) > 0 {
		if len(text) <= limit {
			parts = append(parts, text)
			break
		}

		cut := limit
		for cut > 0 && text[cut] != '\n' {
			cut--
		}
		if cut == 0 {
			cut = limit
		}

		parts = append(parts, text[:cut])
		text = text[cut:]
	}

	return parts
}
