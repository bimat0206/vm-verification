# syntax=docker/dockerfile:1.4
FROM golang:1.24-alpine AS build

WORKDIR /app

# Copy all source code including vendored dependencies
COPY . ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /main cmd/main.go

FROM public.ecr.aws/lambda/provided:al2-arm64

WORKDIR /

COPY --from=build /main /main

RUN mkdir -p /opt/templates

COPY templates/ /opt/templates/

# Set component name for logging
ENV COMPONENT_NAME="PrepareSystemPrompt"

ENTRYPOINT ["/main"]
