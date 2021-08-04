#!/bin/bash

### cleanup
rm -f loggo-logs.pos loggo-containers-ignore
docker-compose up -d redis

### spin loggo
timeout --preserve-status 5 ./build/loggo --no-log-journald --no-sla-exporter \
--flush-interval-sec=1 --buffer-max-size=25 \
--transport="redis" \
--logs-path="tests/fixtures/pods" --position-file-path="loggo-logs.pos" \
--containers-ignore-file-path="loggo-containers-ignore" \
&& echo "Loggo write launch ok" || echo "Loggo write launch failed"

sleep 5
### check results
exec ./build/tests --transport="redis"
