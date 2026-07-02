// Утилита genhash генерирует bcrypt-хеш пароля для ручного заполнения
// поля employees.password_hash (например, при создании первого сотрудника).
//
// Использование:
//
//	go run ./cmd/genhash <пароль>
package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("использование: %s <пароль>", os.Args[0])
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("не удалось сгенерировать хеш: %v", err)
	}
	fmt.Println(string(hash))
}
