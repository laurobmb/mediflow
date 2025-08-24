package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time" // Import adicionado
)

// OllamaService implementa a interface AIService usando uma API Ollama local.
type OllamaService struct {
	apiURL    string
	modelName string
	client    *http.Client // Cliente HTTP agora faz parte da struct
}

// OllamaRequest é a estrutura do payload para a API Ollama.
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse é a estrutura da resposta da API Ollama.
type OllamaResponse struct {
	Response string `json:"response"`
}

// NewOllamaService cria uma nova instância do serviço Ollama.
func NewOllamaService(apiURL, modelName string) *OllamaService {
	if apiURL == "" {
		apiURL = "http://localhost:11434/api/generate" // URL Padrão
	}
	if modelName == "" {
		modelName = "llama3" // Modelo padrão
	}
	return &OllamaService{
		apiURL:    apiURL,
		modelName: modelName,
		// --- ALTERAÇÃO PRINCIPAL AQUI ---
		// Cria um cliente HTTP com um timeout de 30 segundos.
		client: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

// GenerateSummary implementa o método da interface AIService para o Ollama.
func (s *OllamaService) GenerateSummary(ctx context.Context, history string) (string, error) {
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

	// Monta o corpo da requisição
	requestPayload := OllamaRequest{
		Model:  s.modelName,
		Prompt: prompt,
		Stream: false,
	}

	payloadBytes, err := json.Marshal(requestPayload)
	if err != nil {
		log.Printf("Erro ao serializar payload para Ollama: %v", err)
		return "", fmt.Errorf("erro interno ao preparar requisição para IA")
	}

	// Cria e envia a requisição HTTP
	req, err := http.NewRequestWithContext(ctx, "POST", s.apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Erro ao criar requisição para Ollama: %v", err)
		return "", fmt.Errorf("erro interno ao criar requisição para IA")
	}
	req.Header.Set("Content-Type", "application/json")

	// Usa o cliente com timeout definido na struct
	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Erro ao enviar requisição para Ollama: %v", err)
		return "", fmt.Errorf("não foi possível conectar ao serviço de IA local (Ollama)")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler resposta do Ollama: %v", err)
		return "", fmt.Errorf("erro ao ler resposta da IA")
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Ollama retornou status não-OK: %s. Resposta: %s", resp.Status, string(body))
		return "", fmt.Errorf("o serviço de IA local (Ollama) retornou um erro")
	}

	// Decodifica a resposta JSON
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		log.Printf("Erro ao decodificar resposta JSON do Ollama: %v", err)
		return "", fmt.Errorf("resposta inválida do serviço de IA")
	}

	return ollamaResp.Response, nil
}
