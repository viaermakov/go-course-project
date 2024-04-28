# Используем bin в текущей директории для установки плагинов protoc
LOCAL_BIN:=$(CURDIR)/bin

# Добавляем bin в текущей директории в PATH при запуске protoc
PROTOC = PATH="$$PATH:$(LOCAL_BIN)" protoc


.PHONY: .vendor-rm
.vendor-rm:
	rm -rf vendor-proto


# Устанавливаем proto описания google/protobuf
vendor-proto/google/protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf vendor-proto/protobuf &&\
	cd vendor-proto/protobuf &&\
	git sparse-checkout set --no-cone src/google/protobuf &&\
	git checkout
	mkdir -p vendor-proto/google
	mv vendor-proto/protobuf/src/google/protobuf vendor-proto/google
	rm -rf vendor-proto/protobuf


# Устанавливаем proto описания validate
vendor-proto/validate:
	git clone -b main --single-branch --depth=2 --filter=tree:0 \
		https://github.com/bufbuild/protoc-gen-validate vendor-proto/tmp && \
		cd vendor-proto/tmp && \
		git sparse-checkout set --no-cone validate &&\
		git checkout
		mkdir -p vendor-proto/validate
		mv vendor-proto/tmp/validate vendor-proto/
		rm -rf vendor-proto/tmp


# Устанавливаем proto описания google/googleapis
vendor-proto/google/api:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/googleapis/googleapis vendor-proto/googleapis && \
 	cd vendor-proto/googleapis && \
	git sparse-checkout set --no-cone google/api && \
	git checkout
	mkdir -p  vendor-proto/google
	mv vendor-proto/googleapis/google/api vendor-proto/google
	rm -rf vendor-proto/googleapis

vendor-proto/protoc-gen-openapiv2/options:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/grpc-ecosystem/grpc-gateway vendor-proto/grpc-ecosystem && \
 	cd vendor-proto/grpc-ecosystem && \
	git sparse-checkout set --no-cone protoc-gen-openapiv2/options && \
	git checkout
	mkdir -p vendor-proto/protoc-gen-openapiv2
	mv vendor-proto/grpc-ecosystem/protoc-gen-openapiv2/options vendor-proto/protoc-gen-openapiv2
	rm -rf vendor-proto/grpc-ecosystem

# Вендоринг внешних proto файлов
.vendor-proto: .vendor-rm  vendor-proto/google/protobuf vendor-proto/validate vendor-proto/google/api vendor-proto/protoc-gen-openapiv2/options

.PHONY: .bin-deps
.bin-deps:
	$(info Installing binary dependencies...)

	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
    GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0 && \
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@v1.0.4

NOTES_PROTO_PATH:="api/order/v1"

PHONY: .protoc-generate
.protoc-generate: .bin-deps .vendor-proto
	mkdir -p pkg/api/order/v1
	protoc \
	-I ${NOTES_PROTO_PATH} \
	-I vendor-proto \
	--plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go \
	--go_out pkg/api/order/v1 \
	--go_opt paths=source_relative \
	--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc \
	--go-grpc_out pkg/api/order/v1 \
	--go-grpc_opt paths=source_relative \
	--plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate \
	--validate_out="lang=go,paths=source_relative:pkg/api/order/v1" \
	api/order/v1/order.proto
	go mod tidy

run:
	@go run ./cmd/cart-server

build:
	@go build -o bin/cart-server ./cmd/cart-server

test:
	go test -count=1 -short ./... -covermode count

e2e-test:
	go test -count=1 ./tests -tags=integration

benchmark:
	 go test -count=1 -short ./... -bench=. -run=^$ -benchmem -benchtime=100000x

lint:
	gocyclo -ignore ".*_test.*" .
	gocognit -ignore ".*_test.*" .

cover:
	go test -short -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

race:
	go test -v -race ./...

parallel:
	go test -v -parallel 45 -race ./...

gen:
	minimock -i ./internals/service.ProductProvider -o ./internals/service
	minimock -i ./internals/service.CartStorage -o ./internals/service
	minimock -i ./internals/service.LomsProvider -o ./internals/service
