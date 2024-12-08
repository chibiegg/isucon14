package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var erroredUpstream = errors.New("errored upstream")

type paymentGatewayPostPaymentRequest struct {
	Amount int `json:"amount"`
}

type paymentGatewayGetPaymentsResponseOne struct {
	Amount int    `json:"amount"`
	Status string `json:"status"`
}

func requestPaymentGatewayPostPayment(ctx context.Context, paymentGatewayURL string, token string, param *paymentGatewayPostPaymentRequest, retrieveRidesOrderByCreatedAtAsc func() ([]Ride, error)) error {
	b, err := json.Marshal(param)
	if err != nil {
		return err
	}

	// 失敗したらとりあえずリトライ
	// FIXME: 社内決済マイクロサービスのインフラに異常が発生していて、同時にたくさんリクエストすると変なことになる可能性あり
	retry := 0
	idempotencyKey := uuid.NewString()
	for {
		err := func() error {
			// 決済リクエストを投げる
			// Idempotency-Keyを指定しているので同一内容を複数回投げても大丈夫

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, paymentGatewayURL+"/payments", bytes.NewBuffer(b))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Idempotency-Key", idempotencyKey)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusNoContent {
				return fmt.Errorf("failed to execute payment API")
			}

			// NoContentだったら決済が正常に完了しているので終了
			return nil
		}()
		if err != nil {
			if retry < 5 {
				retry++
				time.Sleep(100 * time.Millisecond)
				continue
			} else {
				return err
			}
		}
		break
	}

	return nil
}
