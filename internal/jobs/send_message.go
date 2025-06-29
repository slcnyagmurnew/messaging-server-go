package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"messaging-server/internal/db"
	"messaging-server/internal/logger"
	"messaging-server/internal/model"
	"net/http"
	"os"
	"time"
)

// deliverMessage send each message to webhook, return the result
func deliverMessage(msg model.Message) (*model.RedisMessage, error) {
	webhookURL := os.Getenv("WEBHOOK_URL")
	body, err := json.Marshal(
		model.SentMessage{
			Content: msg.Content,
			Phone:   msg.Phone,
		})

	if err != nil {
		logger.Sugar.Errorf("error marshaling JSON: %v\n", err)
		return nil, err
	}

	// get sending time when request sent
	now := time.Now().Format(time.RFC3339)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))

	if resp.StatusCode != http.StatusAccepted {
		logger.Sugar.Errorf("request could not be accepted error code: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if err != nil {
		logger.Sugar.Errorf("error sending request: %v\n", err)
		return nil, err
	}

	defer resp.Body.Close()

	// read the body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Sugar.Errorf("failed to read webhook response body: %v", err)
		return nil, err
	}

	// set redis message as id and sending time
	var rm model.RedisMessage
	if err = json.Unmarshal(bodyBytes, &rm); err != nil {
		logger.Sugar.Errorf("failed to parse response JSON: %v; body=%s", err, string(bodyBytes))
		return nil, err
	}

	rm.SendingTime = now

	return &rm, nil
}

// SendMessages fetch and send messages
func SendMessages(ctx context.Context) {
	logger.Sugar.Info("messaging job started")

	//// open it to debug graceful shutdown
	//time.Sleep(20 * time.Second)

	// fetch data from postgres
	msgs, err := db.PostgresConnection.FetchUnsentMessages(2)
	if err != nil {
		logger.Sugar.Errorf("messages can not fetched: %v ", err)
		return
	}

	if len(msgs) == 0 {
		logger.Sugar.Warn("no new messages.")
		return
	}

	for _, m := range msgs {
		// send request to webhook
		response, err := deliverMessage(m)
		if err != nil {
			logger.Sugar.Errorf("message %d can not sent: %v ", m.Id, err)
		} else {
			// try to update status of sent message
			err = db.PostgresConnection.UpdateSentMessage(m)
			if err != nil {
				logger.Sugar.Errorf("message %d can not updated: %v ", m.Id, err)
			} else {
				// when successfully updated in postgres, try to cache it
				err = db.RedisConnection.InsertMessage(model.RedisMessage{
					MessageID:   response.MessageID,
					SendingTime: response.SendingTime,
				})
				if err != nil {
					// log and continue to loop
					logger.Sugar.Errorf("message %d can not inserted: %v ", m.Id, err)
				}
			}
		}
	}
}
