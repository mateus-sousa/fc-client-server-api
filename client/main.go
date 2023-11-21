package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type ResponseCotacao struct {
	CotacaoAtual string `json:"cotacao_atual"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Println(err.Error())
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		panic(err)
	}
	if res.StatusCode != http.StatusOK {
		err = errors.New("server url not found")
		log.Println(err)
		panic(err)
	}
	var responseCotacao ResponseCotacao
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err.Error())
		panic(err)
	}
	err = json.Unmarshal(body, &responseCotacao)
	if err != nil {
		log.Println(err.Error())
		panic(err)
	}
	f, err := os.OpenFile("cotacao.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("Dólar: %s\n", responseCotacao.CotacaoAtual))
}
