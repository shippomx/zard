FROM nexus-dev-image.fulltrust.link/base-images/golang:{{.Version}}alpine AS builder

{{if .Chinese}}ENV GOPROXY https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
{{end}}{{if .HasTimezone}}
RUN apk update --no-cache && apk add --no-cache tzdata
{{end}}
WORKDIR /gate/src
ADD . /gate/src

COPY go.mod go.sum ./
RUN go mod download
COPY . .

{{if .Argument}}COPY {{.GoRelPath}}/etc /app/etc
{{end}}RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main {{.GoMainFrom}}


FROM {{.BaseImage}}
WORKDIR /app
{{if .HasTimezone}}COPY --from=builder /usr/share/zoneinfo/{{.Timezone}} /usr/share/zoneinfo/{{.Timezone}}
ENV TZ {{.Timezone}}
{{end}}
COPY --from=builder /gate/src/main /app{{if .Argument}}
COPY --from=builder /app/etc /app/etc{{end}}{{if .HasPort}}
# 服务端口
EXPOSE {{.Port}}
{{end}}{{if .HasMetrics}}
# metrics 端口 对应接口 /metrics /ping
EXPOSE {{.MetricsPort}}
{{end}}
# 如果有grpc端口请取消下面的注释
# EXPOSE 8088
CMD ["/app/main"]
