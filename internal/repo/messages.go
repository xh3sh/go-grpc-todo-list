package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	pb "github.com/xh3sh/go-grpc-todo-list/proto/todo"
)

const (
	TodoTTL = 15 * time.Minute
)

type TodoRepository struct {
	client *redis.Client
}

func NewTodoRepository(client *redis.Client) *TodoRepository {
	return &TodoRepository{
		client: client,
	}
}

func (r *TodoRepository) makeKey(id string) string {
	return fmt.Sprintf("todo:%s", id)
}

func (r *TodoRepository) makeUserKey(userID string) string {
	return fmt.Sprintf("user:%s:todos", userID)
}

func (r *TodoRepository) Save(ctx context.Context, userID string, t *pb.Todo) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, r.makeKey(t.Id), data, TodoTTL)
	pipe.SAdd(ctx, r.makeUserKey(userID), t.Id)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *TodoRepository) Get(ctx context.Context, id string) (*pb.Todo, error) {
	data, err := r.client.Get(ctx, r.makeKey(id)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var t pb.Todo
	if err := json.Unmarshal([]byte(data), &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TodoRepository) List(ctx context.Context, userID string) ([]*pb.Todo, error) {
	ids, err := r.client.SMembers(ctx, r.makeUserKey(userID)).Result()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []*pb.Todo{}, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = r.makeKey(id)
	}

	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var todos []*pb.Todo
	var expiredIDs []interface{}

	for i, val := range vals {
		if val == nil {
			expiredIDs = append(expiredIDs, ids[i])
			continue
		}

		var t pb.Todo
		strVal, ok := val.(string)
		if !ok {
			continue
		}

		if err := json.Unmarshal([]byte(strVal), &t); err != nil {
			continue
		}
		todos = append(todos, &t)
	}

	if len(expiredIDs) > 0 {
		go r.client.SRem(context.Background(), r.makeUserKey(userID), expiredIDs...).Err()
	}

	return todos, nil
}

func (r *TodoRepository) Delete(ctx context.Context, userID string, id string) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, r.makeKey(id))
	pipe.SRem(ctx, r.makeUserKey(userID), id)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *TodoRepository) Update(ctx context.Context, t *pb.Todo) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.makeKey(t.Id), data, TodoTTL).Err()
}
