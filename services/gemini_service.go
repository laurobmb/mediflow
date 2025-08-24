package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiService implementa a interface AIService usando a API do Google Gemini.
type GeminiService struct {
	apiKey    string
	modelName string
}

// NewGeminiService cria uma nova instância do serviço Gemini.
func NewGeminiService(apiKey, modelName string) *GeminiService {
	if modelName == "" {
		modelName = "gemini-1.5-flash-latest" // Modelo padrão
	}
	return &GeminiService{
		apiKey:    apiKey,
		modelName: modelName,
	}
}

// GenerateSummary implementa o método da interface AIService.
func (s *GeminiService) GenerateSummary(ctx context.Context, history string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("a chave de API do Gemini não foi configurada")
	}

	prompt := `Você é um assistente de IA para profissionais de saúde mental. Baseado no histórico de sessões a seguir, gere um resumo conciso e neutro para o terapeuta.

REGRAS IMPORTANTES:
1. NÃO forneça diagnósticos.
2. NÃO sugira tratamentos ou ações.
3. Seja estritamente objetivo e neutro, baseando-se apenas nos dados fornecidos.
4. O objetivo é identificar padrões, evoluções e temas recorrentes.

Estruture o resumo em seções curtas com bullets points, usando markdown:
- **Temas Recorrentes:**
- **Evolução dos Níveis Emocionais:**
- **Pontos de Destaque da Última Sessão:**

HISTÓRICO:
` + history

	client, err := genai.NewClient(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		log.Printf("Erro ao criar cliente Gemini: %v", err)
		return "", fmt.Errorf("falha na configuração do serviço de IA")
	}
	defer client.Close()

	model := client.GenerativeModel(s.modelName)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Erro ao gerar conteúdo com Gemini: %v", err)
		return "", fmt.Errorf("falha ao gerar o resumo de IA")
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if summary, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(summary), nil
		}
	}

	return "A IA não conseguiu gerar um resumo para os dados fornecidos.", nil
}