# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS build

WORKDIR /app

# Copy everything at once - this ensures shared modules are copied before go mod download
COPY . .

# Build with go.work
RUN go build -o /main cmd/main.go

FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

COPY --from=build /main /main

RUN mkdir -p /opt/templates

COPY templates/ /opt/templates/

ENTRYPOINT ["/main"]
