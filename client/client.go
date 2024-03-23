package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Println("Erro ao criar a requisição HTTP:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro ao fazer a requisição HTTP:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Código de Status da Resposta da API:", resp.Status)
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Requisição retornou um código de status diferente de 200 OK")
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Erro ao fazer o parse do JSON:", err)
		return
	}

	bid, ok := result["bid"].(string)
	if !ok {
		fmt.Println("Campo 'bid' não encontrado no JSON")
		return
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Println("Erro ao criar o arquivo cotacao.txt:", err)
		return
	}
	defer file.Close()

	_, err = io.WriteString(file, fmt.Sprintf("Dólar: %s\n", bid))
	if err != nil {
		fmt.Println("Erro ao escrever no arquivo cotacao.txt:", err)
		return
	}

	fmt.Printf("Cotação do Dólar: %s\n", bid)
}
