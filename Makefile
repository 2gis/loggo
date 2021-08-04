PROJECT_PKGS := $$(go list -mod vendor ./... | grep -v /vendor/)

lint:
	golangci-lint run

test:
	echo "mode: count" > coverage-all.out
	for pkg in $(PROJECT_PKGS); do \
		GO111MODULE=on go test -mod vendor -coverprofile=coverage.out -covermode=count -v $$pkg || exit 1 ;\
		if [ -f coverage.out ]; then \
			tail -n +2 coverage.out >> coverage-all.out ;\
			rm coverage.out ;\
		fi; \
	done
	go tool cover -func=coverage-all.out
	rm coverage-all.out

functional-test-amqp: cleanup-docker
	tests/functional_test_amqp.sh

functional-test-redis: cleanup-docker
	tests/functional_test_redis.sh

functional-test-sla: cleanup-docker
	tests/functional_test_sla.sh

cleanup-docker:
	docker-compose stop
	docker-compose rm -f

build:
	mkdir -p build
	GO111MODULE=on CGO_ENABLED=1 GOOS=linux go build -mod vendor -tags netgo -installsuffix cgo -o build/loggo cmd/loggo/main.go

build-test:
	GO111MODULE=on go build -mod vendor -o build/tests cmd/tests/main.go

.PHONY: build test functional-test lint
