package store

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
)

func (db *Store) Init() {
	db.name = "Postgres"
	var err error
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s/%s", os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		log.Fatalf("%v: Ошибка иницциализации: %s\n", db.name, err)
	}

	db.pool, err = pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("%v: не удалось подключиться к БД: %v\n", db.name, err)
	}
	log.Printf("%v: подключено к БД \n", db.name)
}
