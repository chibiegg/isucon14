package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
)

func runMatching() {
	ctx := context.Background()
	// MEMO: 一旦最も待たせているリクエストに適当な空いている椅子マッチさせる実装とする。おそらくもっといい方法があるはず…
	ride := &Ride{}
	if err := db.GetContext(ctx, ride, `SELECT * FROM rides WHERE chair_id IS NULL ORDER BY created_at LIMIT 1`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("no rides", "error", err)
			return
		}
		slog.Error("match error 1", "error", err)
		return
	}

	matched := &Chair{}
	if err := db.GetContext(ctx, matched, "SELECT * FROM chairs INNER JOIN (SELECT id FROM chairs WHERE is_active = TRUE AND is_free = TRUE ORDER BY RAND() LIMIT 1) AS tmp ON chairs.id = tmp.id LIMIT 1"); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("no chairs", "error", err)
			return
		} else {
			slog.Error("match error 2", "error", err)
			return
		}
	}

	if _, err := db.ExecContext(ctx, "UPDATE rides SET chair_id = ? WHERE id = ?", matched.ID, ride.ID); err != nil {
		slog.Error("failed to update ride", "error", err)
		return
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE chairs SET is_free = 0 WHERE id = ?`,
		matched.ID); err != nil {
		slog.Error("failed to update chairs", "error", err)
		return
	}
}
