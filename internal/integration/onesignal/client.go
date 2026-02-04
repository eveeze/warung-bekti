package onesignal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	AppID      string
	APIKey     string
	Limit      int
	Offset     int
	httpClient *http.Client
}

func NewClient(appID, apiKey string) *Client {
	return &Client{
		AppID:      appID,
		APIKey:     apiKey,
		Limit:      50,
		Offset:     0,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type NotificationRequest struct {
	AppID            string            `json:"app_id"`
	IncludePlayerIDs []string          `json:"include_player_ids,omitempty"`
	IncludedSegments []string          `json:"included_segments,omitempty"`
	Headings         map[string]string `json:"headings"`
	Contents         map[string]string `json:"contents"`
	Data             interface{}       `json:"data,omitempty"`
}

func (c *Client) SendNotification(title, message string, playerIDs []string, data interface{}) error {
	reqBody := NotificationRequest{
		AppID:    c.AppID,
		Headings: map[string]string{"en": title},
		Contents: map[string]string{"en": message},
		Data:     data,
	}

	if len(playerIDs) > 0 {
		reqBody.IncludePlayerIDs = playerIDs
	} else {
		reqBody.IncludedSegments = []string{"All"}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://onesignal.com/api/v1/notifications", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Basic "+c.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("onesignal api error: status %d", resp.StatusCode)
	}

	return nil
}
