#!/bin/sh
export GOPROXY=https://goproxy.cn,direct
go install github.com/swaggo/swag/cmd/swag@latest
export PATH=$PATH:~/go/bin
cd /src
go mod tidy
swag init -g ./cmd/api-server/main.go -o ./docs/swagger --parseDependency --parseInternal
cat ./docs/swagger/swagger.json