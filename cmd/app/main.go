package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"github.com/xh3sh/go-grpc-todo-list/internal/db"
	"github.com/xh3sh/go-grpc-todo-list/internal/repo"
	"github.com/xh3sh/go-grpc-todo-list/internal/todo"
	pb "github.com/xh3sh/go-grpc-todo-list/proto/todo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const swaggerJSONPath = "proto/openapiv2/todo/todo.swagger.json"

var (
	grpcAddr = flag.String("grpc", ":50051", "gRPC listen address")
	httpAddr = flag.String("http", ":80", "HTTP listen address")
)

func main() {
	flag.Parse()
	if err := godotenv.Load(); err != nil {
		log.Println("env файл не найден, используются стандартные параметры")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb, err := db.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к Redis: %v", err)
	}
	ctx := context.Background()

	todoRepo := repo.NewTodoRepository(rdb)

	// 1. Запуск gRPC сервера
	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	// Экземпляр todo-сервера для gRPC и HTTP
	todoServer := todo.NewServer(todoRepo)

	pb.RegisterTodoServiceServer(grpcServer, todoServer)
	go func() {
		log.Printf("gRPC server listening on %s", *grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC serve: %v", err)
		}
	}()

	// 2. HTTP Gateway
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Основной мультиплексор
	rootMux := http.NewServeMux()

	// Мультиплексор для gRPC Gateway
	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := pb.RegisterTodoServiceHandlerFromEndpoint(ctx, gwMux, *grpcAddr, opts); err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	// // Регистрируем gRPC Gateway обработчики
	rootMux.Handle("/api/", http.StripPrefix("/api", gwMux))
	rootMux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, swaggerJSONPath)
	})

	// Обработчики для рендера html/template
	rootMux.HandleFunc("/", todoServer.HandleTodosPage)
	rootMux.HandleFunc("/htmx/todo/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			todoServer.HandleGetTodo(w, r)
		case http.MethodPost:
			todoServer.HandleCreateTodo(w, r)
		case http.MethodPatch:
			todoServer.HandlePatchTodo(w, r)
		case http.MethodDelete:
			todoServer.HandleDeleteTodo(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Обработчик для статики
	rootMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 3. Запуск HTTP сервера
	log.Printf("HTTP gateway listening on %s", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, todoServer.AuthMiddleware(rootMux)); err != nil {
		log.Fatalf("HTTP serve: %v", err)
	}
}
