PROTO_DIR    := proto
TODO_OUT     := proto
GOOGLEAPIS   := third_party/googleapis
GW_APIS      := third_party/grpc-gateway

default: gen up

.PHONY: gen
gen:
	protoc \
	  -I$(PROTO_DIR) \
	  -I$(GOOGLEAPIS) \
	  -I$(GW_APIS) \
	  --go_out=paths=source_relative:$(TODO_OUT) \
	  --go-grpc_out=paths=source_relative:$(TODO_OUT) \
	  --grpc-gateway_out=paths=source_relative:$(TODO_OUT) \
	  --openapiv2_out=$(TODO_OUT)/openapiv2 \
	  --openapiv2_opt logtostderr=true \
	  $(PROTO_DIR)/todo/*.proto

up:
	docker network create projects-network || true
	docker compose up -d --build