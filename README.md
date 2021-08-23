# loggo

Loggo was initialy developed to address a need of simple log agent for a Kubernetes cluster.

It's supposed to run as a container on cluster nodes, with knowledge about some kubernetes entities and existing logging engines implementations.

### Build
Loggo requires `libsystem-dev` package libraries to be installed in build environment due to it's journald dependency.
```
sudo apt install libsystemd-dev
make build
```

### Usage with K8S Daemonset
Loggo can be used as a container with K8S Daemonset.
```
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: {{ app_name }}
  labels:
    k8s-app: {{ app_name }}
    kubernetes.io/cluster-service: "true"
spec:
  selector:
    matchLabels:
      k8s-app: {{ app_name }}
  template:
    metadata:
      labels:
        k8s-app: {{ app_name }}
      annotations:
        prometheus.io/path: "/metrics"
        prometheus.io/port: "8080"
        prometheus.io/scrape: "true"
    spec:
      serviceAccountName: {{ app_name }}
      priorityClassName: system-node-critical
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: kubernetes.io/os
                operator: NotIn
                values:
                - windows
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      - key: CriticalAddonsOnly
        operator: Exists
      containers:
      - name: {{ app_name }}
        image: docker-hub.2gis.ru/2gis-io/loggo:{{ image_version }}
        volumeMounts:

        ports:
          - containerPort: 8080

        env:
        - name: DC
          value: "{{ dc }}"
        - name: PURPOSE
          value: "{{ purpose }}"
        - name: LOGSTASH_PREFIX
          value: "{{ logstash_prefix }}"

        - name: NODE_HOSTNAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName

        - name: TRANSPORT
          value: "{{ transport }}"
        - name: BUFFER_SIZE_MAX
          value: "{{ buffer_size_max }}"
        - name: READER_BUFFER_SIZE
          value: "{{ reader_buffer_size }}"

        {%- if transport == "redis" %}
        - name: REDIS_HOSTNAME
          value: "{{ redis_hostname }}"
        - name: REDIS_KEY
          value: "{{ redis_key }}"
        {%- endif %}

        {%- if transport == "amqp" %}
        - name: AMQP_URL
          value: "{{ amqp_url }}"
        - name: AMQP_EXCHANGE
          value: "{{ amqp_exchange }}"
        - name: AMQP_ROUTING_KEY
          value: "{{ amqp_routing_key }}"
        {%- endif %}

        - name: LOGS_PATH
          value: "{{ logs_path }}"
        - name: POSITION_FILE_PATH
          value: "{{ position_file_path }}"
        - name: FROM_TAIL_FLAG
          value: "{{ from_tail_flag }}"

        - name: READ_RATE_RULES_PATH
          value: "{{ read_rate_rules_path }}"
        - name: LOG_JOURNALD
          value: "{{ log_journald }}"
        - name: JOURNALD_PATH
          value: "{{ journald_path }}"
        - name: MACHINE_ID_PATH
          value: "{{ machine_id_path }}"

        - name: SLA_EXPORTER
          value: "{{ sla_exporter }}"
        - name: SLA_BUCKETS
          value: "{{ sla_buckets }}"
        - name: SERVICE_DEFAULT_DOMAIN
          value: "{{ service_default_domain }}"
        - name: SERVICE_UPDATE_INTERVAL_SEC
          value: "{{ service_update_interval_sec }}"
        - name: METRICS_RESET_INTERVAL_SEC
          value: "{{ metrics_reset_interval_sec }}"

        - name: LOG_LEVEL
          value: "{{ log_level }}"

        volumeMounts:
        - name: rates-config-volume
          mountPath: "/configuration"
        - name: varlog
          mountPath: /var/log
        - name: nodemachineid
          mountPath: "{{ machine_id_path }}"
        - name: varlibdockercontainers
          mountPath: /var/lib/docker/containers
          readOnly: true
      terminationGracePeriodSeconds: 30
      
      volumes:
      - name: rates-config-volume
        configMap:
          name: {{ app_name }}
      - name: varlog
        hostPath:
          path: /var/log
      - name: varlibdockercontainers
        hostPath:
          path: /var/lib/docker/containers
      - name: nodemachineid
        hostPath:
          path: /etc/machine-id
```
Configuration parameters can be set via CLI or in environment variables as shown above.

## Logic
Basically, Loggo reads path specified (`LOGS_PATH`), assuming that it's the directory with Docker/CRI logs and metafiles that are created by container engine in Kubernetes.

Once log file with corresponding metadata is found, Loggo starts to read it, considering its content as "container logs" and does the following:
1. Parses lines according to container engine format, constructing basic entry map. In case of nested json log structure, inner maps and lists get flattened with dot notation (see below).
2. Enriches entry map with K8S-related metadata, such as container name, pod, namespace, etc. After that, entry map consists of the original log message fields, and additional fields that Loggo adds.
3. If SLI gathering logic is enabled, tries to read fixed nginx access log fields to construct several Prometheus metrics from logs.
4. Serializes the resulting flat entry map to JSON dictionary string and adds it to a batch. Batches are being sent to the transport (redis, etc).

### Expected log formats
Loggo is purposed for reading log files generated by Docker and CRI/Containerd container engines.

### Docker engine
Node paths `/var/log/pods` contains symlinks and `/var/lib/docker/containers` contains log and `config.v2.json` files. The latter is used to get k8s container metadata and state.

Therefore, both `/var/log` and `/var/lib/docker/containers` directories must be volumes to expose their content to Loggo container that runs on this node. 

Symlinks are being searched in LOGS_PATH (usually `var/log/pods`). They get resolved,
and the K8S related metadata are being taken from config.v2.json file of that container.

Log line is the json with "log" field, that should contain map containing user log.

### CRI/Containerd
With CRI/Containerd, `/var/log/pods` path contains log files themselves, and the container metadata is being
taken from log file path.

Log line is a plain string of known format.

### Log fields transformations
Let's say we have docker engine log string:
```
{"log":"{ \"sla\":             false, \"remote_addr\":             \"10.154.18.198\", \"remote_user\":             \"\", \"time_local\":              \"09/Jan/2018:12:07:48 +0700\", \"time_msec\":               \"1515474468.396\",\"request_method\": \"POST\", \"server_protocol\": \"HTTP/1.1\", \"request_uri\":                 \"/api/mis/getlocation\", \"status\":                   200, \"host\":                    \"api.2gis.com\", \"request_time\":             0.010, \"upstream_response_time\":  \"0.078\", \"body_bytes_sent\":          213, \"http_referer\":            \"https://cerebro.2gis.test/\", \"http_user_agent\":         \"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:57.0) Gecko/20100101 Firefox/57.0\", \"request_id\":              \"d010f6631c81407596a7ff19aaf6312b\", \"geoip.location\":          \"0,0\", \"upstream_request_id\":     \"\" }\n","stream":"stderr","time":"2018-01-09T05:08:03.100481875Z"}
```
It would be successfuly parsed and the resulting JSON string (linted for readme) that would go to the transport will be:
```
{
   "body_bytes_sent":213,
   "dc":"n3",
   "docker.container_id":"id",
   "geoip.location":"0,0",
   "host":"api.2gis.com",
   "http_referer":"https://cerebro.2gis.test/",
   "http_user_agent":"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:57.0) Gecko/20100101 Firefox/57.0",
   "kubernetes.container_name":"container",
   "kubernetes.namespace_name":"namespace",
   "kubernetes.node_hostname":"localhost",
   "kubernetes.pod_name":"pod",
   "logstash_prefix":"k8s-unknown",
   "namespace":"namespace",
   "purpose":"staging",
   "remote_addr":"10.154.18.198",
   "remote_user":"",
   "request_id":"d010f6631c81407596a7ff19aaf6312b",
   "request_method":"POST",
   "request_time":0.01,
   "request_uri":"/api/mis/getlocation",
   "server_protocol":"HTTP/1.1",
   "status":200,
   "stream":"stdout",
   "time":"2018-01-09T05:08:03.100481875Z",
   "time_local":"09/Jan/2018:12:07:48 +0700",
   "time_msec":"1515474468.396",
   "type":"containers",
   "upstream_request_id":"",
   "upstream_response_time":"0.020, 0.078",
   "upstream_response_time_float":0.078,
   "upstream_response_time_total":0.098
}
```
As mentioned, Loggo flattens the original nested log message structure: if it contains maps or strings, we would have them flattened to one-level dictionary with dot notation for field names: `a:{b: "test"}` to `a.b:"test"` and so on.

In addition, loggo adds its own fields to every entry. These are of two types:
* Fields that are set from env/CLI loggo configuration (see --help), such as `dc`, `purpose`, `logstash_prefix`,
`kubernetes.node_hostname` and others.
* Fields with fixed names, values of which are taken from container info. In the example above they are:
```
 "kubernetes.container_name":"container",
 "kubernetes.namespace_name":"namespace",
 "kubernetes.pod_name":"pod",
```

given that "container", "namespace" and "pod" are the container, namespace and pod names respectively, as read from container's `config.v2.json` file or containerd container path.
### Reserved fields
Some field names are considered service ones and **are removed** from the record after processing.
Incomplete list of those fields:
* `logging` with bool value. Present `logging: false` (boolean) prevents the entry from being sent to the transport.
* `sla` with bool value. Present `sla: true` serves as a mark for log entry further parsing as SLI.

Though this is a backward compatibility logic, currently users are advised to avoid using these 
fields in their log messages for the other purposes.

### Reading system journal (journald)
In order to read system journal, it's needed to expose journald path (usually `/var/log/journal`) and machineid (`/etc/node-machine-id`) path to the loggo container.

System journal reading can be disabled by specifying `--no-log-journald` CLI flag.

### Limiting (throttling) reading speed
In some cases it may be useful to limit the logs reading speed to prevent high pressure on logs receiver.
With Loggo one can do it by specifying rate limit rules file.
This file expected to be yaml of the following form:
```yaml
rates:
    - namespace: "kubernetes_namespace_0"
      rate: 100000 # lines per second
    - namespace: "kubernetes_namespace_1"
      pod: "router-perf.*"
      rate: 1000
    - namespace: ".*"
      rate: 1000
```
Rules are being checked in the following order:
1. Both pod and namespace specified.
2. Pod name specified.
3. Namespace specified.

The rules of the same type are checked one by one until the first match.

If a pod that matches the rule produces logs faster than specified in `rate` field, reading of its log file slows down to limit. 

The fact of such throttling is registered in Prometheus counter `container_throttling_delay_seconds_total`.
If no rules file specified, the default value is taken from `read-rate-default` CLI flag, `READ_RATE_DEFAULT` env with default of 1000.

### SLI gathering
TBD.