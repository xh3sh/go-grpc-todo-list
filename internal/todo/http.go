package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	pb "github.com/xh3sh/go-grpc-todo-list/proto/todo"
)

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseFiles(
		filepath.Join("views", "index.html"),
		filepath.Join("views", "todos.html"),
	)
	if err != nil {
		log.Fatalf("ошибка парсинга шаблонов: %v", err)
	}
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			if c, err := r.Cookie("X-User-ID"); err == nil {
				userID = c.Value
			}
		}

		if userID != "" {
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (s *Server) HandleTodosPage(w http.ResponseWriter, r *http.Request) {
	userID := s.getUserID(r.Context())
	todos, err := s.repo.List(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list todos: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		filepath.Join("views", "index.html"),
		filepath.Join("views", "todos.html"),
	)
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Todos []*pb.Todo
	}{
		Todos: todos,
	}

	sort.Slice(data.Todos, func(i, j int) bool {
		if data.Todos[i].Done != data.Todos[j].Done {
			return !data.Todos[i].Done
		}

		id1, err1 := strconv.ParseInt(data.Todos[i].Id, 10, 64)
		id2, err2 := strconv.ParseInt(data.Todos[j].Id, 10, 64)
		if err1 != nil || err2 != nil {
			return false
		}
		return id1 < id2
	})

	err = tmpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) HandleGetTodo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	todo, err := s.GetTodo(r.Context(), &pb.GetRequest{Id: id})
	if err != nil {
		http.Error(w, "не найдено: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "todo-item", todo); err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) HandleCreateTodo(w http.ResponseWriter, r *http.Request) {
	var in pb.Todo
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		http.Error(w, "невалидный JSON", http.StatusBadRequest)
		return
	}

	todo, err := s.CreateTodo(r.Context(), &in)
	if err != nil {
		http.Error(w, "ошибка создания: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = templates.ExecuteTemplate(w, "todo-item", todo)
	if err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) HandlePatchTodo(w http.ResponseWriter, r *http.Request) {
	var in pb.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		fmt.Println(err)
		http.Error(w, "невалидный JSON", http.StatusBadRequest)
		return
	}

	todo, err := s.UpdateTodo(r.Context(), &in)
	if err != nil {
		http.Error(w, "ошибка обновления: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "todo-item", todo); err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) HandleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/htmx/todo/")
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "неверный ID", http.StatusBadRequest)
		return
	}

	_, err := s.DeleteTodo(r.Context(), &pb.DeleteRequest{Id: id})
	if err != nil {
		http.Error(w, "ошибка удаления: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
