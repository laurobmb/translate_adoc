package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath" // NOVO: para manipular caminhos de arquivos
	"strings"       // NOVO: para manipular o nome do arquivo

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/google/generative-ai-go/genai"
)

// NOVO: Define o nome do diretório de traduções como uma constante
const traducoesDir = "traducoes"

func main() {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		fmt.Println("A variável de ambiente GOOGLE_API_KEY não está definida.")
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("Uso: go run tradutor.go <caminho_para_o_documento.txt>")
		return
	}
	docPath := os.Args[1]

	content, err := os.ReadFile(docPath)
	if err != nil {
		fmt.Printf("Erro ao ler o arquivo de entrada: %v\n", err)
		return
	}

	// --- INÍCIO DAS ALTERAÇÕES ---

	// 1. Criar o diretório de traduções, se ele não existir.
	//    os.MkdirAll é como 'mkdir -p', não retorna erro se o diretório já existe.
	if err := os.MkdirAll(traducoesDir, 0755); err != nil {
		fmt.Printf("Erro ao criar o diretório '%s': %v\n", traducoesDir, err)
		return
	}

	// 2. Criar o nome do arquivo de saída.
	//    Exemplo: "documento.txt" -> "traducoes/documento_traduzido.txt"
	originalFileName := filepath.Base(docPath)
	fileExtension := filepath.Ext(originalFileName)
	baseName := strings.TrimSuffix(originalFileName, fileExtension)
	outputFileName := fmt.Sprintf("%s_traduzido%s", baseName, fileExtension)
	outputPath := filepath.Join(traducoesDir, outputFileName)

	// 3. Criar o arquivo de saída onde a tradução será salva.
	outputFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Erro ao criar o arquivo de saída '%s': %v\n", outputPath, err)
		return
	}
	defer outputFile.Close() // Garante que o arquivo será fechado ao final da função.

	// --- FIM DAS ALTERAÇÕES ---

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		fmt.Printf("Erro ao criar o cliente: %v\n", err)
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	prompt := fmt.Sprintf("Traduza o seguinte texto para o português do Brasil:\n\n%s", string(content))

	iter := model.GenerateContentStream(ctx, genai.Text(prompt))

	fmt.Printf("Traduzindo '%s' e salvando em '%s'...\n", docPath, outputPath)

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Erro ao gerar conteúdo: %v\n", err)
			return
		}

		for _, cand := range resp.Candidates {
			if cand.Content != nil {
				for _, part := range cand.Content.Parts {
					if txt, ok := part.(genai.Text); ok {
						// ALTERADO: Em vez de imprimir na tela, escreve no arquivo.
						if _, err := outputFile.WriteString(string(txt)); err != nil {
							fmt.Printf("Erro ao escrever no arquivo de saída: %v\n", err)
							return
						}
					}
				}
			}
		}
	}

	// NOVO: Mensagem de sucesso
	fmt.Println("Tradução concluída com sucesso!")
}