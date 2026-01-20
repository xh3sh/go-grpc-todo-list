package todo

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/xh3sh/go-grpc-todo-list/internal/repo"
	pb "github.com/xh3sh/go-grpc-todo-list/proto/todo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

type Server struct {
	pb.UnimplementedTodoServiceServer
	repo *repo.TodoRepository
}

func NewServer(r *repo.TodoRepository) *Server {
	return &Server{
		repo: r,
	}
}

func (s *Server) getUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok && id != "" {
		return id
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		vals := md.Get("x-user-id")
		if len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}

	return "default_user"
}

func (s *Server) CreateTodo(ctx context.Context, t *pb.Todo) (*pb.Todo, error) {
	if t.Title == "" {
		return nil, fmt.Errorf("название задачи не может быть пустым")
	}
	if utf8.RuneCountInString(t.Description) > 270 {
		return nil, fmt.Errorf("описание задачи слишком длинное")
	}

	if t.Id == "" {
		t.Id = generateID()
	}
	t.Date = generateDate()

	userID := s.getUserID(ctx)
	if err := s.repo.Save(ctx, userID, t); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save todo: %v", err)
	}

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
	t, err := s.repo.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get todo: %v", err)
	}
	if t == nil {
		return nil, status.Errorf(codes.NotFound, "todo %q not found", req.Id)
	}
	return t, nil
}

func (s *Server) ListTodos(_ *pb.Empty, stream pb.TodoService_ListTodosServer) error {
	ctx := stream.Context()
	userID := s.getUserID(ctx)

	todos, err := s.repo.List(ctx, userID)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to list todos: %v", err)
	}

	for _, t := range todos {
		if err := stream.Send(t); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) DeleteTodo(ctx context.Context, req *pb.DeleteRequest) (*pb.Empty, error) {
	userID := s.getUserID(ctx)
	if err := s.repo.Delete(ctx, userID, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete todo: %v", err)
	}
	return &pb.Empty{}, nil
}

func (s *Server) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.Todo, error) {
	current, err := s.repo.Get(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch todo: %v", err)
	}
	if current == nil {
		return nil, status.Errorf(codes.NotFound, "todo %q not found", req.Id)
	}

	if req.Title != "" {
		current.Title = req.Title
	}
	if req.Description != "" {
		current.Description = req.Description
	}

	current.Done = req.Done

	if err := s.repo.Update(ctx, current); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update todo: %v", err)
	}

	return current, nil
}
