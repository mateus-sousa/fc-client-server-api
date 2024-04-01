package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"time"
)

type CotacaoAtual struct {
	Usdbrl USBBRL `json:"USDBRL"`
}

type USBBRL struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type OutputCotacao struct {
	CotacaoAtual string `json:"cotacao_atual"`
}

type CotacaoHandler struct {
	db *sql.DB
}

func NewCotacaoHandler(db *sql.DB) *CotacaoHandler {
	return &CotacaoHandler{db: db}
}

func main() {
	db, err := sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	handler := NewCotacaoHandler(db)
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", handler.GetCotacao)
	log.Println("server listening in 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func (h *CotacaoHandler) GetCotacao(w http.ResponseWriter, r *http.Request) {
	log.Println("request received")
	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	body, err := io.ReadAll(res.Body)
	var cotacaoAtual CotacaoAtual
	err = json.Unmarshal(body, &cotacaoAtual)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	err = SaveRequest(ctx, h.db, cotacaoAtual)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(OutputCotacao{CotacaoAtual: cotacaoAtual.Usdbrl.Bid})
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func SaveRequest(ctx context.Context, db *sql.DB, cotacaoAtual CotacaoAtual) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	id := uuid.New().String()
	createdAt := time.Now().Unix()
	_, err := db.ExecContext(ctx, "insert into requests (id, bid, created_at) values($1, $2, $3)", id, cotacaoAtual.Usdbrl.Bid, createdAt)
	if err != nil {
		return err
	}
	return nil
}
