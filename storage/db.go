package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// NewDBConnection cria e retorna uma conexão com o banco de dados PostgreSQL.
func NewDBConnection() (*sql.DB, error) {
	dbType := os.Getenv("DB_TYPE")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	// Adiciona um valor padrão para a porta se não for especificada
	if dbPort == "" {
		dbPort = "5432"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open(dbType, connStr)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir a conexão com o banco de dados: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("falha ao pingar o banco de dados: %w", err)
	}

	fmt.Println("Conexão com o banco de dados estabelecida com sucesso!")
	return db, nil
}