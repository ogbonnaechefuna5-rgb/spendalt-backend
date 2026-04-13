package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$14$TiGwTZ9elzBsbWjBmmlFzOb5ATT9qVq.QG9TGhO143jCHcLYwP.YW"
	password := "test1234"
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Println("MISMATCH:", err)
	} else {
		fmt.Println("MATCH")
	}
}
