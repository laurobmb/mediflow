package handlers

import (
	"regexp"
	"strconv"
)

// IsCPFValid realiza a validação completa de um CPF, incluindo os dígitos verificadores.
func IsCPFValid(cpf string) bool {
	// 1. Limpa o CPF, deixando apenas os números
	re := regexp.MustCompile(`[^0-9]`)
	cpf = re.ReplaceAllString(cpf, "")

	// 2. Verifica se o CPF tem 11 dígitos
	if len(cpf) != 11 {
		return false
	}

	// 3. Verifica se todos os dígitos são iguais (ex: 111.111.111-11), que são inválidos
	allSame := true
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	// 4. Cálculo do primeiro dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (10 - i)
	}
	remainder := sum % 11
	dv1 := 0
	if remainder >= 2 {
		dv1 = 11 - remainder
	}
	
	// 5. Compara o dígito calculado com o dígito real
	realDv1, _ := strconv.Atoi(string(cpf[9]))
	if dv1 != realDv1 {
		return false
	}

	// 6. Cálculo do segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (11 - i)
	}
	remainder = sum % 11
	dv2 := 0
	if remainder >= 2 {
		dv2 = 11 - remainder
	}

	// 7. Compara o segundo dígito calculado com o dígito real
	realDv2, _ := strconv.Atoi(string(cpf[10]))
	if dv2 != realDv2 {
		return false
	}

	// Se passou por todas as verificações, o CPF é válido
	return true
}