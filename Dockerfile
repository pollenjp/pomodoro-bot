FROM golang:1.19-bullseye as builder
COPY . /src
WORKDIR /src
RUN go build -ldflags="-w -s -linkmode external -extldflags -static" -o ./build/app ./main.go

FROM scratch as prod
COPY --from=builder /src/build/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/app"]
