## syntax=docker/dockerfile:1
#
#FROM golang:1.18-alpine
#
## Set destination for COPY
#WORKDIR /app
#
## Download Go modules
#COPY go.mod go.sum ./
#RUN go mod download
#
## Copy the source code. Note the slash at the end, as explained in
## https://docs.docker.com/engine/reference/builder/#copy
#COPY *.go ./
#
## Build
#RUN CGO_ENABLED=0 GOOS=linux go build -o /bizhubBackend
#
## Optional:
## To bind to a TCP port, runtime parameters must be supplied to the docker command.
## But we can document in the Dockerfile what ports
## the application is going to listen on by default.
## https://docs.docker.com/engine/reference/builder/#expose
#EXPOSE 8080
#
## Run
#CMD ["/bizhubBackend"]


# Start from the official Go image
FROM golang:1.18-alpine
# Set the Current Working Directory inside the container
WORKDIR /app
# Copy go.mod and go.sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
# Copy the source from the current directory to the Working Directory inside the container
COPY . .
# Set environment variable for the views directory
ENV VIEWS_DIR=/app/internal/views
# Build the Go app
RUN go build -o /app/main /app/main.go
# Expose port 8080 to the outside world
EXPOSE 3000
# Set environment variable for Gin mode
ENV GIN_MODE=release
# Run the executable
CMD ["/app/main"]