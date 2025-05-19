package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		db: db,
	}
}

type Cotacao struct {
	ID          int
	Code        string    `json:"code"`
	CodeIn      string    `json:"codein"`
	Name        string    `json:"name"`
	High        string    `json:"high"`
	Low         string    `json:"low"`
	VarBid      string    `json:"varBid"`
	PctChange   string    `json:"pctChange"`
	Bid         string    `json:"bid"`
	Ask         string    `json:"ask"`
	TimeStamp   string    `json:"timestamp"`
	CreatedDate time.Time `json:"create_date"`
}

func (s *Server) CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Println("Erro ao encontrar pagina")
		return
	}

	cotacao, err := BuscaCotacao()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Erro ao acessar API: %v", err)
		return
	}

	err = s.SalvarBanco(cotacao)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Erro ao acessar DB: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao)
}

// Função responsavel por dar Get só no bid
func GetBid(ctx context.Context) (*string, error) {

	select {
	case <-time.After(300 * time.Millisecond):
		log.Println("Bid pego com sucesso")
		cotacao, err := BuscaCotacao()
		if err != nil {
			fmt.Println("Erro ao buscar bid")
			return nil, err
		}

		bid := cotacao.Bid
		return &bid, nil

	case <-ctx.Done():
		fmt.Println("Timeout atingido")
		return nil, ctx.Err()
	}
}

// Função responsavel por consumir a API
func BuscaCotacao() (*Cotacao, error) {

	ctx := context.Background()
	log.Println("Request Iniciada")
	defer log.Println("Request finalizada")
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Println("Erro ao fazer requisição")
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Erro ao obter resposta")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Erro ao ler o body")
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %v", err)
	}

	cotacaoData, ok := data["USDBRL"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected JSON structure")
	}

	cotacao := &Cotacao{
		Code:        cotacaoData["code"].(string),
		CodeIn:      cotacaoData["codein"].(string),
		Name:        cotacaoData["name"].(string),
		High:        cotacaoData["high"].(string),
		Low:         cotacaoData["low"].(string),
		VarBid:      cotacaoData["varBid"].(string),
		PctChange:   cotacaoData["pctChange"].(string),
		Bid:         cotacaoData["bid"].(string),
		Ask:         cotacaoData["ask"].(string),
		TimeStamp:   cotacaoData["timestamp"].(string),
		CreatedDate: time.Now(),
	}

	return cotacao, nil
}

// Função responsavel por salvar a cotação no banco de dados
func (s *Server) SalvarBanco(cotacao *Cotacao) error {

	ctx := context.Background()
	log.Println("Request Iniciada")
	defer log.Println("Request finalizada")
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()

	query := "INSERT INTO cotacao(code, code_in, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) VALUES(?,?,?,?,?,?,?,?,?,?,?)"

	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		log.Println("Erro ao preparar query")
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, cotacao.Code, cotacao.CodeIn, cotacao.Name, cotacao.High, cotacao.Low,
		cotacao.VarBid, cotacao.PctChange, cotacao.Bid, cotacao.Ask, cotacao.TimeStamp, cotacao.CreatedDate)
	if err != nil {
		log.Println("Erro ao inserir no banco")
		return err
	}

	return nil
}
