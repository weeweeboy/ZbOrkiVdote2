package main

import (
	"ZbOrkiVdote/internal/application/user/usecases"
	"ZbOrkiVdote/internal/infrastructure/storage"
	"ZbOrkiVdote/internal/presentation/handlers"
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	logger := slog.Default()
	logger.Info("Логгер подключен")

	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp4", "77.88.8.8:53")
		},
	}

	bot, err := tgbotapi.NewBotAPIWithClient(
		"8966745932:AAFJ47qH7irJseyqFS_IRpYmt7X0yGgRdjc",
		tgbotapi.APIEndpoint,
		&http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					d := net.Dialer{}

					if strings.HasSuffix(addr, "api.telegram.org:443") {
						addr = "149.154.167.220:443"
					}

					return d.DialContext(ctx, "tcp4", addr)
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}

	logger.Info("Бот создан")
	bot.Debug = true

	newStorage, err := storage.NewStorage("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", logger)
	if err != nil {
		panic(err)
	}

	serviceLayer := usecases.NewUserUsecase(newStorage, logger)
	handler := handlers.NewHandlers(bot, serviceLayer, logger)
	handler.Start(context.Background())
}
