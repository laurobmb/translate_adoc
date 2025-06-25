package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// A função getTranslation permanece a mesma
func getTranslation(ctx context.Context, model *genai.GenerativeModel, contentToTranslate string) (string, error) {
	prompt := fmt.Sprintf("Traduza o seguinte texto para o português do Brasil, mantendo a formatação original do AsciiDoc o mais fielmente possível:\n\n%s", contentToTranslate)
	
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("erro ao gerar conteúdo da API: %w", err)
	}

	var translatedText strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					translatedText.WriteString(string(txt))
				}
			}
		}
	}

	if translatedText.Len() == 0 {
		return "", fmt.Errorf("a tradução retornou um texto vazio")
	}

	return translatedText.String(), nil
}

func main() {
	ctx := context.Background()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		fmt.Println("ERRO: A variável de ambiente GOOGLE_API_KEY não está definida.")
		return
	}

	// ALTERAÇÃO AQUI: LÓGICA DE BUSCA RECURSIVA
	fmt.Println("Procurando por arquivos .adoc em todos os subdiretórios...")
	var adocFiles []string
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // Propaga erros (ex: permissão negada para ler um dir)
		}
		// Verifica se é um arquivo (não um diretório) e se termina com .adoc
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".adoc") {
			adocFiles = append(adocFiles, path)
		}
		return nil // Retorna nil para continuar a caminhada
	})

	if err != nil {
		fmt.Printf("ERRO: Ocorreu um erro ao procurar pelos arquivos: %v\n", err)
		return
	}
	// FIM DA ALTERAÇÃO

	if len(adocFiles) == 0 {
		fmt.Println("Nenhum arquivo .adoc encontrado.")
		return
	}

	fmt.Printf("Encontrados %d arquivos .adoc para processar.\n\n", len(adocFiles))

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		fmt.Printf("ERRO: Falha ao criar o cliente da API: %v\n", err)
		return
	}
	defer client.Close()
	model := client.GenerativeModel("gemini-1.5-flash")

	// O resto do loop de processamento permanece idêntico
	for _, originalPath := range adocFiles {
		fmt.Printf("--- Processando arquivo: %s ---\n", originalPath)

		backupPath := originalPath + ".bkp"
		err := os.Rename(originalPath, backupPath)
		if err != nil {
			fmt.Printf("  -> ERRO: Falha ao criar backup para '%s'. Pulando para o próximo. Erro: %v\n\n", originalPath, err)
			continue
		}
		fmt.Printf("  -> Backup criado: %s\n", backupPath)

		content, err := os.ReadFile(backupPath)
		if err != nil {
			fmt.Printf("  -> ERRO: Falha ao ler o arquivo de backup '%s'. Pulando para o próximo. Erro: %v\n\n", backupPath, err)
			continue
		}
		fmt.Println("  -> Conteúdo original lido com sucesso.")

		fmt.Println("  -> Enviando para tradução...")
		translatedContent, err := getTranslation(ctx, model, string(content))
		if err != nil {
			fmt.Printf("  -> ERRO: Falha na tradução para '%s'. Pulando para o próximo. Erro: %v\n\n", originalPath, err)
			continue
		}
		fmt.Println("  -> Tradução recebida.")

		err = os.WriteFile(originalPath, []byte(translatedContent), 0666)
		if err != nil {
			fmt.Printf("  -> ERRO: Falha ao salvar o arquivo traduzido '%s'. Pulando para o próximo. Erro: %v\n\n", originalPath, err)
			continue
		}
		fmt.Printf("  -> Tradução salva com sucesso em: %s\n\n", originalPath)
	}

	fmt.Println("--- Processo concluído! ---")
}