package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	db, err := sql.Open("sqlite3", "file:a.db?mode=memory")
	if err != nil {
		slog.ErrorContext(ctx, "failed to open db", slog.Any("error", err))
		os.Exit(1)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute)

	// log DBStats per 1 second
	go func() {
		ticker := time.Tick(time.Second)
		for {
			select {
			case <-ticker:
				slog.InfoContext(ctx, "stats", slog.Any("stats", db.Stats()))
			case <-ctx.Done():
				return
			}
		}
	}()

	for range 20 {
		go func() {
			ticker2 := time.Tick(10 * time.Nanosecond)
			for {
				select {
				case <-ticker2:
					var discard int
					r := db.QueryRowContext(ctx, "SELECT 1")
					_ = r.Scan(&discard)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	<-ctx.Done()
	db.Close()
}
