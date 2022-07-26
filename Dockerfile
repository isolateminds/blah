FROM golang:1.18

WORKDIR /usr/src/

COPY go.mod go.sum ./
RUN go mod download && go mod verify

ARG ENV_NAME
ARG PORT

ENV ENV_FILE=.${ENV_NAME}.env

COPY . .

EXPOSE $PORT
RUN go build -v -o /usr/local/bin/blah ./cmd/server/main.go

CMD ["blah"]