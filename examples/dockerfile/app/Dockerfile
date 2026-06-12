FROM golang:alpine
RUN apk add --no-cache wget
WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go build -o server .
EXPOSE 8080
CMD ["/app/server"]