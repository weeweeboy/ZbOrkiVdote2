package seed

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

// делаем структурки и их методы для того, чтобы в будущем создать интефейс
type Item struct {
	ItemID   int    `json:"id"`
	ItemName string `json:"dname"`
}

func (i Item) ID() int      { return i.ItemID }
func (i Item) Name() string { return i.ItemName }

type Hero struct {
	DisplayName string `json:"displayName"`
	HeroID      int    `json:"hero_id"`
}

func (h Hero) ID() int      { return h.HeroID }
func (h Hero) Name() string { return h.DisplayName }

// вот и он
type Row interface {
	ID() int
	Name() string
}

// создаем функцию которая создает таблицу с переданным названием
func createTable(ctx context.Context, conn *pgx.Conn, logger *slog.Logger, tableName string) error {
	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s(
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL)`, tableName)

	_, err := conn.Exec(ctx, query)
	if err != nil {
		logger.Error("Unable to create table",
			slog.String("table", tableName),
			slog.Any("error", err))
	}
	return err
}

// функция с дженериком для того, чтобы не писать отдельную для предметов и героев. только при вызове кода функция будет знать
// что она парсит и куда
func fetchAndDecode[T any](client *http.Client, url string) (T, error) {
	var result T

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("error decoding JSON: %w", err)
	}

	return result, nil
}

// вставляет данные полученные после парсинга в таблицу. она также не знает вставляет она героев либо предметы!
// все благодаря дженерикам и интерфейсу
func insertAll[T Row](ctx context.Context, conn *pgx.Conn, logger *slog.Logger, table string, rows []T) {
	for _, r := range rows {
		if _, err := conn.Exec(ctx,
			fmt.Sprintf("INSERT INTO %s (id, name) VALUES ($1, $2) ON CONFLICT DO NOTHING", table),
			r.ID(), r.Name()); err != nil {

			logger.Error("Unable to insert row",
				slog.String("table", table),
				slog.Int("id", r.ID()),
				slog.Any("error", err))
		}
	}
}

func main() {
	logger := slog.Default()

	conn, err := pgx.Connect(context.Background(), "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		logger.Error("Unable to connect to database", slog.Any("error", err))
		return
	}
	defer conn.Close(context.Background())

	logger.Info("connected to postgresql://localhost:5432/postgres")

	// перебор названий, чтоб не писать много
	for _, table := range []string{"heroes", "items"} {
		if err := createTable(context.Background(), conn, logger, table); err != nil {
			return
		}
	}

	// создаем клиент для запросов
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false,
			},
		},
		Timeout: 30 * time.Second,
	}
	// в переменную собираем все данные о предметах. items это объект класса Item (функция об этом не знала, но мы ей указали)
	items, err := fetchAndDecode[map[string]Item](client, "https://raw.githubusercontent.com/odota/dotaconstants/master/build/items.json")
	if err != nil {
		logger.Error("Failed to fetch items", slog.Any("error", err))
		return
	}
	logger.Info("Successfully fetched items", slog.Int("count", len(items)))

	// создает слайс этих самых айтемов и заносим их туда
	itemSlice := make([]Item, 0, len(items))
	for _, v := range items {
		itemSlice = append(itemSlice, v)
	}

	// вставляем все айтемы
	insertAll(context.Background(), conn, logger, "items", itemSlice)

	// здесь все аналогично
	heroes, err := fetchAndDecode[[]Hero](client, "https://dota2protracker.com/api/heroes/list")
	if err != nil {
		logger.Error("Failed to fetch heroes", slog.Any("error", err))
		return
	}
	logger.Info("Successfully fetched heroes", slog.Int("count", len(heroes)))

	insertAll(context.Background(), conn, logger, "heroes", heroes)

	logger.Info("Successfully populated tables")
}
