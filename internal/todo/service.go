package todo

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	pb "github.com/xh3sh/go-grpc-todo-list/proto/todo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const cleanupTime = 15

type Server struct {
	pb.UnimplementedTodoServiceServer
	mu          sync.Mutex
	todos       map[string]*pb.Todo
	stopCleanup chan struct{}
}

func NewServer() *Server {
	s := &Server{
		todos:       make(map[string]*pb.Todo),
		stopCleanup: make(chan struct{}),
	}

	go s.startCleanupTimer(cleanupTime * time.Minute)

	return s
}

func (s *Server) CreateTodo(ctx context.Context, t *pb.Todo) (*pb.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.Id == "" {
		t.Id = generateID()
	}
	if t.Title == "" {
		return nil, fmt.Errorf("название задачи не может быть пустым")
	}
	if utf8.RuneCountInString(t.Description) > 270 {
		return nil, fmt.Errorf("описание задачи слишком длинное")
	}
	t.Date = generateDate()
	s.todos[t.Id] = t
	return t, nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func generateDate() string {
	t := time.Now()
	dateTimeStr := t.Format("2006-01-02 15:04:05")
	return dateTimeStr
}

func (s *Server) GetTodo(ctx context.Context, req *pb.GetRequest) (*pb.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.todos[req.Id]; ok {
		return t, nil
	}
	return nil, status.Errorf(codes.NotFound, "todo %q not found", req.Id)
}

func (s *Server) ListTodos(_ *pb.Empty, stream pb.TodoService_ListTodosServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.todos {
		if err := stream.Send(t); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) DeleteTodo(ctx context.Context, req *pb.DeleteRequest) (*pb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.todos, req.Id)
	return &pb.Empty{}, nil
}

func (s *Server) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.todos[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "todo %q not found", req.Id)
	}
	t.Done = req.Done
	return t, nil
}
