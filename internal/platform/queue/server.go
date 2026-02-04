package queue

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
)

type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func NewServer(redisAddr string, redisPassword string, concurrency int) *Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("ERROR: Task %s failed: %v", task.Type(), err)
			}),
		},
	)

	return &Server{
		server: srv,
		mux:    asynq.NewServeMux(),
	}
}

func (s *Server) Handle(pattern string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Run() error {
	return s.server.Run(s.mux)
}

func (s *Server) Start() error {
	return s.server.Start(s.mux)
}

func (s *Server) Stop() {
	s.server.Stop()
}
