FROM golang:1.16 as builder
EXPOSE 8080
COPY . /src
RUN cd /src; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o blog ./cmd/blog

FROM scratch
COPY --from=builder /src/blog .
ENTRYPOINT ["/blog", "serve"]
