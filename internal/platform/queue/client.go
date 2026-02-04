package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

type Client struct {
	client *asynq.Client
}

func NewClient(redisAddr string, redisPassword string) *Client {
	return &Client{
		client: asynq.NewClient(asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		}),
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) EnqueueLowStockAlert(payload PayloadLowStock) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeLowStockAlert, data)
	_, err = c.client.Enqueue(task, asynq.ProcessIn(0)) // Immediate
	return err
}

func (c *Client) EnqueueNewTransaction(payload PayloadNewTransaction) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	task := asynq.NewTask(TypeNewTransaction, data)
	_, err = c.client.Enqueue(task)
	return err
}

// Generic Enqueue for specialized needs or simple forwarding
func (c *Client) EnqueueSendNotification(payload PayloadNotificationSend) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	task := asynq.NewTask(TypeNotificationSend, data)
	_, err = c.client.Enqueue(task)
	return err
}
