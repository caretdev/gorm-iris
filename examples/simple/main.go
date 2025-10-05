package main

import (
	"fmt"

	iris "github.com/caretdev/gorm-iris"
	"gorm.io/gorm"
)

type User struct {
  ID    int
  Name  string
  Email string
}

func main() {
  dsn := "iris://_SYSTEM:SYS@localhost:1972/USER"
  db, err := gorm.Open(iris.Open(dsn), &gorm.Config{})
  if err != nil {
    panic("failed to connect to IRIS")
  }

	// Auto-migrate schema
  db.AutoMigrate(&User{})

  // Create
  db.Create(&[]User{
		{Name: "Johh", Email: "john@example.com"},
		{Name: "Johh1", Email: "john1@example.com"},
		{Name: "John2", Email: "john2@example.com"},
	})

  // Read
  var user User
  db.First(&user, "email = ?", "john1@example.com")
	fmt.Printf("Found: ID: %d; Name: %s\n", user.ID, user.Name)
}
