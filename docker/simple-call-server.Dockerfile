FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /server ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /server /server
ENTRYPOINT ["/server"]
