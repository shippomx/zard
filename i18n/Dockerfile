FROM nexus-dev-image.fulltrust.link/base-images/golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o i18n-tools tools/main.go 




FROM nexus-dev-image.fulltrust.link/base-images/alpine:latest

RUN apk add --no-cache git bash

WORKDIR /app

COPY --from=builder /app/i18n-tools .
COPY --from=builder /app/push.sh .
RUN chmod +x ./push.sh

ENTRYPOINT ["/app/push.sh"]