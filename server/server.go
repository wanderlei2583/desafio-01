package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		fmt.Println("Erro ao abrir o banco de dados:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT,
		dataHora TIMESTAMP
	)`)
	if err != nil {
		fmt.Println("Erro ao criar a tabela no banco de dados:", err)
		return
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			http.Error(w, "Erro ao criar a requisição HTTP", http.StatusInternalServerError)
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Erro ao fazer a requisição HTTP", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("Requisição retornou código de status %d", resp.StatusCode), http.StatusInternalServerError)
			return
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Println("Erro ao fazer o parse do JSON:", err)
			http.Error(w, "Erro ao fazer o parse do JSON", http.StatusInternalServerError)
			return
		}

		log.Println("Resposta JSON da API:", result)

		bid, ok := result["USDBRL"].(map[string]interface{})["bid"].(string)
		if !ok {
			log.Println("Campo 'bid' não encontrado no JSON")
			http.Error(w, "Campo 'bid' não encontrado no JSON", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO cotacoes (bid, dataHora) VALUES (?, ?)", bid, time.Now())
		if err != nil {
			fmt.Println("Erro ao inserir cotação no banco de dados:", err)
		}

		cotacao := map[string]string{"bid": bid}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cotacao)
	})

	http.ListenAndServe(":8080", nil)
}
