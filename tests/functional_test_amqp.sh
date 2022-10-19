#!/bin/bash

### cleanup
rm -f loggo-logs.pos
docker-compose up -d rabbit

### prepare rabbit
./build/tests --create-rabbit-queues

### spin loggo
timeout --preserve-status 5 ./build/loggo/loggo --no-log-journald  --no-sla-exporter \
--flush-interval-sec=1 --buffer-max-size=2 \
--position-file-path="loggo-logs.pos" --logs-path="tests/fixtures/pods" && echo "ok" || echo "bad"

### check results
exec ./build/tests
