package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/guiyones/desafio-client-server-api.git/server"
)

func GetBidFromAPI() *string {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	bid, err := server.GetBid(ctx)
	if err != nil {
		log.Printf("Erro ao chamar API")
		return nil
	}

	fmt.Println(string(*bid))

	return bid
}

func CriarArquivo() {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
	}
	defer file.Close()

	bid := string(*GetBidFromAPI())

	_, err = file.WriteString(fmt.Sprintf("Dólar:{%s}", bid))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
	}

	fmt.Println("Arquivo criado com sucesso!")
}
