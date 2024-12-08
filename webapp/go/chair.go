package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrNoChairs = fmt.Errorf("no chairs")

func getChairNotification(ctx context.Context, chair *Chair) (*chairGetNotificationResponseData, error) {
	status := ""
	tx, err := db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	ride := &Ride{}
	yetSentRideStatus := RideStatus{}

	if err := tx.GetContext(ctx, ride, `SELECT * FROM rides WHERE chair_id = ? ORDER BY updated_at DESC LIMIT 1`, chair.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoChairs
		}
		return nil, err
	}

	if err := tx.GetContext(ctx, &yetSentRideStatus, `SELECT * FROM ride_statuses WHERE ride_id = ? AND chair_sent_at IS NULL ORDER BY created_at ASC LIMIT 1`, ride.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			status, err = getLatestRideStatus(ctx, tx, ride.ID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		status = yetSentRideStatus.Status
	}

	user := &User{}
	err = tx.GetContext(ctx, user, "SELECT * FROM users WHERE id = ? FOR SHARE", ride.UserID)
	if err != nil {
		return nil, err
	}

	if yetSentRideStatus.ID != "" {
		_, err := tx.ExecContext(ctx, `UPDATE ride_statuses SET chair_sent_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, yetSentRideStatus.ID)
		if err != nil {
			return nil, err
		}
	}

	if yetSentRideStatus.Status == "COMPLETED" {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE chairs SET is_free = 1 WHERE id = ?`,
			ride.ChairID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	d := chairGetNotificationResponseData{
		RideID: ride.ID,
		User: simpleUser{
			ID:   user.ID,
			Name: fmt.Sprintf("%s %s", user.Firstname, user.Lastname),
		},
		PickupCoordinate: Coordinate{
			Latitude:  ride.PickupLatitude,
			Longitude: ride.PickupLongitude,
		},
		DestinationCoordinate: Coordinate{
			Latitude:  ride.DestinationLatitude,
			Longitude: ride.DestinationLongitude,
		},
		Status: status,
	}
	return &d, nil
}
