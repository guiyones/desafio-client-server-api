package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/guiyones/desafio-client-server-api.git/client"
	"github.com/guiyones/desafio-client-server-api.git/server"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	cotacaoDB := server.NewServer(db)
	client.CriarArquivo()

	http.HandleFunc("/cotacao", cotacaoDB.CotacaoHandler)
	fmt.Println("Server rodando na porta 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}

}
