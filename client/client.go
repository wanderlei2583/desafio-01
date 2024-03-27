package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Quote struct {
	Bid string `json:"bid"`
}

func fetchQuote() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var quote Quote
	if err := json.NewDecoder(resp.Body).Decode(&quote); err != nil {
		return "", err
	}

	return quote.Bid, nil
}

func main() {
	quote, err := fetchQuote()
	if err != nil {
		log.Fatalf("Erro ao Buscar a cotação: %v", err)
	}

	content := []byte("Dólar: " + quote + "\n")
	err = os.WriteFile("cotacao.txt", content, 0644)
	if err != nil {
		log.Fatalf("Erro ao gravar o arquivo: %v", err)
	}
}
