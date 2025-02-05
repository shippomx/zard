Name: xxl-job-executor-sample # 注册到xxl-job-admin的执行器名称用的是这个Name,限制以小写字母开头，由小写字母、数字和中划线组成
ServerAddr: "http://127.0.0.1:8080/xxl-job-admin" # xxl-job-admin的地址
AccessToken: "default_token" # 访问 xxl-job-admin 的token
ExecutorIP: "127.0.0.1" # 可以不配置，默认取本机IP（ipv4.LocalIP()）
ExecutorPort: "8082" # 可以不配置，默认端口为8082

DevServer:
  Enabled: true

Log:
  # 是否启用统计日志
  Stat: false


