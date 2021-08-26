# loggo

Loggo was initialy developed to address a need of simple log agent for a Kubernetes cluster.

It's supposed to run as a container on cluster nodes, with knowledge about some kubernetes entities and existing logging
engines implementations.

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

Basically, Loggo reads path specified (`LOGS_PATH`), assuming that it's the directory with Docker/CRI logs and metafiles
that are created by container engine in Kubernetes.

Once log file with corresponding metadata is found, Loggo starts to read it, considering its content as "container logs"
and does the following:

1. Parses lines according to container engine format, constructing basic entry map. In case of nested json log
   structure, inner maps and lists get flattened with dot notation (see below).
2. Enriches entry map with K8S-related metadata, such as container name, pod, namespace, etc. After that, entry map
   consists of the original log message fields, and additional fields that Loggo adds.
3. If SLI gathering logic is enabled, tries to read fixed nginx access log fields to construct several Prometheus
   metrics from logs.
4. Serializes the resulting flat entry map to JSON dictionary string and adds it to a batch. Batches are being sent to
   the transport (redis, etc).

### Expected log formats

Loggo is purposed for reading log files generated by Docker and CRI/Containerd container engines.

### Docker engine

Node paths `/var/log/pods` contains symlinks and `/var/lib/docker/containers` contains log and `config.v2.json` files.
The latter is used to get k8s container metadata and state.

Therefore, both `/var/log` and `/var/lib/docker/containers` directories must be volumes to expose their content to Loggo
container that runs on this node.

Symlinks are being searched in LOGS_PATH (usually `var/log/pods`). They get resolved, and the K8S related metadata is
being taken from config.v2.json file of that container.

Log line is the json with "log" field, that should contain map containing user log.

### CRI/Containerd

With CRI/Containerd, `/var/log/pods` path contains log files themselves, and the container metadata is being taken from
log file path.

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

As mentioned, Loggo flattens the original nested log message structure: if it contains maps or strings, we would have
them flattened to one-level dictionary with dot notation for field names: `a:{b: "test"}` to `a.b:"test"` and so on.

In addition, loggo adds its own fields to every entry. These are of two types:

* Fields that are set from env/CLI loggo configuration (see --help), such as `dc`, `purpose`, `logstash_prefix`,
  `kubernetes.node_hostname` and others.
* Fields with fixed names, values of which are taken from container info. In the example above they are:

```
 "kubernetes.container_name":"container",
 "kubernetes.namespace_name":"namespace",
 "kubernetes.pod_name":"pod",
```

given that "container", "namespace" and "pod" are the container, namespace and pod names respectively, as read from
container's `config.v2.json` file or containerd container path.

### Reserved fields

Some field names are considered service ones and **are removed** from the record after processing. Incomplete list of
those fields:

* `logging` with bool value. Present `logging: false` (boolean) prevents the entry from being sent to the transport.
* `sla` with bool value. Present `sla: true` serves as a mark for log entry further parsing as SLI.

Though this is a backward compatibility logic, currently users are advised to avoid using these fields in their log
messages for the other purposes.

### Reading system journal (journald)

In order to read system journal, it's needed to expose journald path (usually `/var/log/journal`) and
machineid (`/etc/node-machine-id`) path to the loggo container.

System journal reading can be disabled by specifying `--no-log-journald` CLI flag.

### Limiting (throttling) reading speed

In some cases it may be useful to limit the logs reading speed to prevent high pressure on logs receiver. With Loggo one
can do it by specifying rate limit rules file. This file expected to be yaml of the following form:

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

If a pod that matches the rule produces logs faster than specified in `rate` field, reading of its log file slows down
to limit.

The fact of such throttling is registered in Prometheus counter `container_throttling_delay_seconds_total`. If no rules
file specified, the default value is taken from `read-rate-default` CLI flag, `READ_RATE_DEFAULT` env with default of

1000.

### SLI gathering

While it is considered bad practice to collect metrics from logs, it can sometimes seem like a viable option. Examples
include the need to collect metrics consistently and centrally for all applications in the cluster, and/or the desire to
compare the data of metrics endpoints of your applications with another source. Loggo supports custom logic for
constructing a number of predefined metrics from predefined log fields. Nginx access log format was taken as a basis for
it. When SLI gathering is enabled (`--sla-exporter/SLA_EXPORTER=true`), Loggo needs access to K8S API to get list of
services. Cluster role for this could look like this:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app: {{ app_name }}
  name: {{ app_name }}
rules:
  - apiGroups: [""]
    resources: ["services", "namespaces"]
    verbs: ["get", "list"]
```

Then it's needed to specify names for annotations that would be used: AnnotationExporterEnable, AnnotationExporterPaths,
AnnotationSLADomains. Exact names of those annotations is configurable via CLI/env. The values of annotations with
chosen names should contain "true", a list of paths and comma-separated string of domains respectively (see example
below).

Example of service annotations, given that `AnnotationExporterEnable`, `AnnotationExporterPaths`, `AnnotationSLADomains`
names are left on defaults: `loggo.sla/enable`, `loggo.sla/paths`, and `loggo.sla/domains`.

```
loggo.sla/enable: "true"
loggo.sla/paths: |
  [
    // "path" : "path regexps"
    {"/vt": ["^/vt.*$"]},
    {"/tiles?ts=online_sd": ["^/tiles\\?.*ts=online_sd"]},
    {"/tiles?ts=other": ["^/tiles.*$"]},
    {"/poi": ["^/poi.*$"]},
  ]
loggo.sla/domains: "domainprefixwithoutdots, domain.with.dots"
```

Notice that paths specification is a raw json inside yaml, so slashes should be properly escaped. If one of the domains
specified without dots, it's considered as prefix and gets joined with services default domain, specified via
CLI/env (`default-service-domain`/`SERVICE_DEFAULT_DOMAIN`). Otherwise it's taken as is. The following steps are taken:

1. Loggo gets list of namespaces, for each namespace it gets a list of services with their annotations.
2. It tries to make representation of each service. It requires the following annotations to be present.
3. Matches several required log fields to service paths and domain and increments metrics.

The following fields are required to be present in log message to let it be parsed as SLI message:

* `sla` with value of true (bool)
* `host`
* `request_method`
* `request_uri`
* `server_protocol`
* `request_time`
* `status`

Optional fields for upstream response time related metrics:

* `upstream_pod_name`
* `upstream_response_time`

Then service is searched by `host` (`host` is compared with domains in domains annotation), and `request_uri` is matched
against path regexps. Once match is found, the path from the annotation (for instance, `"/vt"` from example above)
becomes a `path` label value. The metrics that getting updated at this moment are `http_request_count`
, `http_request_total_count` and `http_request_time`.

#### Upstream response time processing

If the `upstream_response_time` field is present in log message, it gets processed the following way:

* If it's a string, the resulting message would include:
    * `upstream_response_time_float` with the last (or single) upstream response time.
    * `upstream_response_time_total` with sum of all comma-separated non-empty (`-`) upstream values.
* If it's a float, it will be taken as is and set to both of fields

The `upstream_response_time_total` value is taken into account in the `http_upstream_response_time_total` metric.

### Metrics

Common:

Name | Type | Labels | Description
| ------ | ------ | --------- | --------- |
| log_message_count | Counter | "namespace", "pod", "container" | Log message count for a container. |
| container_throttling_delay_seconds_total | Counter | "namespace", "pod", "container" | Indicates particular container's total throttle time. |

SLI related (appear only if SLI gathering is enabled and there are messages that are matching with K8S service
annotations):

Name | Type | Labels | Description
| ------ | ------ | --------- | --------- |
| http_request_count | Counter | "method", "service", "path", "status", "protocol", "upstream_pod_name" | The total number of requests processed.|
| http_request_total_count | Counter | "service" | The same, by service, without detailed labels (for messages without `path` info). |
| http_request_time | Histogram | "method", "service", "path", "protocol", "upstream_pod_name" | Histogram for HTTP request time.
| http_upstream_response_time_total | Histogram | "method", "service", "path", "protocol", "upstream_pod_name" | Histogram for HTTP upstream response time. |

Metrics get reset every `--metrics-reset-interval-sec`/`METRICS_RESET_INTERVAL_SEC` to clean the stale containers data
up. Buckets for histogram metrics are configurable with CLI/env `sla-buckets`/`SLA_BUCKETS`. Default setting
is `"0.01 0.02 0.04 0.06 0.08 0.1 0.15 0.2 0.25 0.3 0.4 0.5 0.6 0.7 0.8 0.9 1 1.2 1.5 1.75 2 3 4 5 8 10 20 60"`.
