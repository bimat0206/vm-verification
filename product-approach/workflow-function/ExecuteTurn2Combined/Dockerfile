FROM public.ecr.aws/docker/library/golang:1.24 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /tmp/main ./cmd/main.go

FROM public.ecr.aws/docker/library/alpine:3.18
COPY --from=builder /tmp/main /var/task/main
CMD ["/var/task/main"]
