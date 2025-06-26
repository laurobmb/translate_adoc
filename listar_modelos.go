package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("A variável de ambiente GOOGLE_API_KEY não está definida.")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Println("Buscando modelos disponíveis...")
	fmt.Println("---------------------------------")

	// O método ListModels retorna um iterador para percorrer os modelos
	iter := client.ListModels(ctx)
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// Imprime as informações de cada modelo encontrado
		fmt.Printf("Nome: %s\n", m.Name)
		fmt.Printf("  -> Display Name: %s\n", m.DisplayName)
		fmt.Printf("  -> Descrição: %s\n\n", m.Description)
	}
}