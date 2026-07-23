package parser

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type Sborka struct {
	MatchID  int64  `json:"match_id"`
	Mmr      int    `json:"mmr"`
	HeroID   int    `json:"hero_id"`
	Position string `json:"position"`
	Won      bool   `json:"won"`
	Data     struct {
		Mmr   int    `json:"mmr"`
		Won   int    `json:"won"`
		Name  string `json:"name"`
		Role  string `json:"role"`
		Items []struct {
			Minute int `json:"minute"`
			ItemID int `json:"item_id"`
		} `json:"items"`
		IsPro int `json:"is_pro"`
	} `json:"data"`
}

func ParseSborka(ctx context.Context, url string) ([]Sborka, error) {
	var result []Sborka

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				d := net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}

				if strings.HasSuffix(addr, "dota2protracker.com:443") {
					addr = "104.27.145.95:443"
				}

				return d.DialContext(ctx, "tcp4", addr)
			},
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 15 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false,
			},
		},
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://dota2protracker.com/")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("нет данных для этой комбинации героя и позиции")
		}
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	return result, nil
}
