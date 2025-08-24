package services

import "context"

// AIService é a interface que define o contrato para qualquer provedor de IA.
// Qualquer provedor (Gemini, Ollama, etc.) deve implementar este método.
type AIService interface {
	GenerateSummary(ctx context.Context, history string) (string, error)
}