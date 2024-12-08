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
			slog.Info("no rides")
			return
		}
		slog.Info("match error 1")
		return
	}

	matched := &Chair{}
	if err := db.GetContext(ctx, matched, "SELECT * FROM chairs INNER JOIN (SELECT id FROM chairs WHERE is_active = TRUE AND is_free = TRUE ORDER BY RAND() LIMIT 1) AS tmp ON chairs.id = tmp.id LIMIT 1"); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("no chairs")
			return
		} else {
			slog.Info("match error 2")
			return
		}
	}

	if _, err := db.ExecContext(ctx, "UPDATE rides SET chair_id = ? WHERE id = ?", matched.ID, ride.ID); err != nil {
		slog.Info("failed to update ride")
		return
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE chairs SET is_free = 0 WHERE id = ?`,
		ride.ChairID); err != nil {
		slog.Info("failed to update chairs")
		return
	}
}
