FROM golang:1.23-bullseye

RUN apt update && apt install -y nodejs npm
RUN go install github.com/securego/gosec/v2/cmd/gosec@latest
RUN npm install -g snyk
