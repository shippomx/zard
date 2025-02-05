Name: {{.serviceName}}.rpc
ListenOn: 0.0.0.0:8088
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: {{.serviceName}}.rpc

{{ if .gateway }}
RestConf:
  Name: {{.serviceName}}.rest
  Host: 0.0.0.0
  Port: 8080
{{end}}