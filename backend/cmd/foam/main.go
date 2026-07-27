package main

// @title Foam
// @version 1.0
// @description Foam full-stack development scaffold backend.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Use "Bearer <admin access token>".

import (
	"fmt"
	"os"

	"github.com/Rain-kl/Foam/backend/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
