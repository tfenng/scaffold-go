package main

import (
	"context"
	"fmt"
	"os"

	"scaffold-api/internal/config"
	"scaffold-api/internal/db"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "down") {
		return fmt.Errorf("usage: migrate [up|down]")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	sqlDB, err := db.Open(ctx, cfg)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	switch os.Args[1] {
	case "up":
		return db.MigrateUp(ctx, sqlDB)
	case "down":
		return db.MigrateDown(ctx, sqlDB)
	default:
		return fmt.Errorf("unknown migration direction: %s", os.Args[1])
	}
}
