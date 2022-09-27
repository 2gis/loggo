#!/bin/bash

### cleanup
rm -f loggo-logs.pos loggo-containers-ignore
docker-compose up -d redis

### start metrics endpoint
./build/loggo/loggo --no-log-journald --sla-service-source-path="tests/fixtures/config.yaml" \
--flush-interval-sec=1 --service-update-interval-sec=1 --transport="redis" \
--logs-path="tests/fixtures/pods" --buffer-max-size=2 \
--position-file-path="loggo-logs.pos" --containers-ignore-file-path="loggo-containers-ignore" &
PID=$!
sleep 2

### check results
./build/tests --sla-testing --transport="redis"

RC=$?
kill -s INT $PID
wait $PID
exit $RC
