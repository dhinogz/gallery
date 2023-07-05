package main

import (
	"fmt"

	"github.com/dhinogz/lenslocked/models"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Order struct {
	ID          int
	UserID      int
	Amount      int
	Description int
}

func main() {
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected!")

	us := models.UserService{
		DB: db,
	}
	user, err := us.Create("dav1@test.com", "test123")
	if err != nil {
		panic(err)
	}
	fmt.Println(user)
}
