FROM golang:1.19-bullseye as builder
COPY . /src
WORKDIR /src
RUN go build -ldflags="-w -s" -o ./build/app ./main.go

FROM scratch as prod
COPY --from=builder /src/build/app /app
CMD ["/app"]
