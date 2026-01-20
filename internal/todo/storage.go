package todo

import (
	"time"

	pb "github.com/xh3sh/go-grpc-todo-list/proto/todo"
)

func (s *Server) startCleanupTimer(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.clearTodos()
		case <-s.stopCleanup:
			return
		}
	}
}

func (s *Server) clearTodos() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.todos = make(map[string]*pb.Todo)
}
