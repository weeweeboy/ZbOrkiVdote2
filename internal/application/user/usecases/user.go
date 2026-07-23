package usecases

import (
	"ZbOrkiVdote/internal/application/user/interfaces/storage"
	"ZbOrkiVdote/internal/infrastructure/parser"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type UserUsecase struct {
	userStorage storage.UserStorage
	logger      *slog.Logger
}

func NewUserUsecase(userStorage storage.UserStorage, logger *slog.Logger) *UserUsecase {
	return &UserUsecase{
		userStorage: userStorage,
		logger:      logger,
	}
}
func (u *UserUsecase) GetSborka(ctx context.Context, name string, position string) (string, error) {
	heroId, err := u.userStorage.GetH(ctx, name)
	if err != nil {
		u.logger.Error("Не смогли найти айдишник героя", slog.Any("error", err))
		return "", err
	}

	url := fmt.Sprintf("https://dota2protracker.com/api/hero-matches?hero_id=%d&position=pos+%s&sort_by=mmr", heroId, position)

	time.Sleep(time.Second)

	matches, err := parser.ParseSborka(ctx, url)
	if err != nil {
		u.logger.Error("Ошибка при парсинге сборок", slog.Any("error", err))
		return "", err
	}

	if len(matches) == 0 {
		return "К сожалению, сборки для этого героя на данной позиции не найдены.", nil
	}

	limit := 5
	if len(matches) < limit {
		limit = len(matches)
	}

	resultText := fmt.Sprintf("📊 **Последние %d сборок для %s (Позиция %s)**\n\n", limit, name, position)

	for i := 0; i < limit; i++ {
		match := matches[i]

		var itemsList string
		for _, item := range match.Data.Items {
			itemName, _ := u.userStorage.GetI(ctx, item.ItemID)
			itemsList += fmt.Sprintf("  • %d мин: Предмет ID %s\n", item.Minute, itemName)
		}
		if itemsList == "" {
			itemsList = "  • Закуп отсутствует\n"
		}

		resultText += fmt.Sprintf(
			"🎮 **Матч #%d** | %s | MMR: %d (%s)\n"+
				"👤 Игрок: %s\n"+
				"📦 **Закуп:**\n%s"+
				"-----------------------------------\n",
			i+1,
			match.Data.Role,
			match.Data.Mmr,
			getWinStatus(match.Won),
			match.Data.Name,
			itemsList,
		)
	}

	return resultText, nil
}

func getWinStatus(won bool) string {
	if won {
		return "Победа 🟢"
	}
	return "Поражение 🔴"
}
