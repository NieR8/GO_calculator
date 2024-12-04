package main

import (
	"github.com/NieR8/myProject/internal/application"
)

func main() {
	app := application.New()
	app.RunServer()
}
