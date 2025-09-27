# gorm-iris

[GORM](https://gorm.io) dialect for [InterSystems IRIS](https://www.intersystems.com/iris/) built on top of [go-irisnative](https://github.com/caretdev/go-irisnative).

> Status: **alpha**. APIs may change. Feedback and PRs welcome.

---

## Installation

```bash
go get github.com/caretdev/gorm-iris
```

---

## DSN format

The DSN is the same as for `go-irisnative`:

```
iris://user:password@host:1972/Namespace
```

Example:

```bash
export IRIS_DSN='iris://_SYSTEM:SYS@localhost:1972/USER'
```

---

## Quick start

```go
package main

import (
    "fmt"
    "gorm.io/gorm"
    "github.com/caretdev/gorm-iris/iris" // import dialect
)

type Person struct {
    ID   int    `gorm:"primaryKey"`
    Name string
}

func main() {
    dsn := "iris://_SYSTEM:SYS@localhost:1972/USER?sslmode=disable"
    db, err := gorm.Open(iris.Open(dsn), &gorm.Config{})
    if err != nil { panic(err) }

    // Auto-migrate schema
    db.AutoMigrate(&Person{})

    // Insert
    db.Create(&Person{ID: 1, Name: "Alice"})

    // Query
    var p Person
    db.First(&p, 1)
    fmt.Println("Found:", p)

    // Update
    db.Model(&p).Update("Name", "Alice Updated")

    // Delete
    db.Delete(&p)
}
```

---

## Features

* ✅ Drop-in GORM support for InterSystems IRIS
* ✅ Schema migration (`AutoMigrate`)
* ✅ CRUD operations
* ✅ Transactions
* ✅ Based on `database/sql` driver (`go-irisnative`)

---

## Compatibility

* Go: 1.21+
* GORM: v2
* InterSystems IRIS: 2022.1+

---

## License

MIT License

---

## Contributing

* Please run `go vet` and `go test ./...` before PRs.
* Document any missing IRIS-specific features or differences.
* Add examples for advanced GORM patterns as tested.
