package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Quote struct {
	Bid string `json:"bid"`
}

type APIResponse struct {
	USDBRL Quote `json:"USDBRL"`
}

func fetchDollarQuote(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	return apiResp.USDBRL.Bid, nil
}

func recordQuoteInDB(ctx context.Context, quote string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	db, err := sql.Open("sqlite3", "quotes.db")
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS quotes (id INTEGER PRIMARY KEY, bid TEXT)")
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, "INSERT INTO quotes (bid) VALUES (?)", quote)
	return err
}

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	quote, err := fetchDollarQuote(ctx)
	if err != nil {
		http.Error(w, "Não foi possível buscar a cotação", http.StatusInternalServerError)
		log.Println("fetchDollarQuote error:", err)
		return
	}

	if err := recordQuoteInDB(ctx, quote); err != nil {
		http.Error(w, "Não foi possível gravar a cotação", http.StatusInternalServerError)
		log.Println("recordQuoteInDB error:", err)
		return
	}

	json.NewEncoder(w).Encode(Quote{Bid: quote})
}

func main() {
	http.HandleFunc("/cotacao", quoteHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
