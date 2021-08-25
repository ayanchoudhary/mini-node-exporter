FROM golang:1.14.3 as builder

ENV GO111MODULE="on"

# Copy the code from the host and compile it
WORKDIR /app

# install dependencies
ADD ./go.sum ./go.sum
ADD ./go.mod ./go.mod

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go


# Expose port 23333 to the outside world
EXPOSE 23333

# Command to run the executable
CMD ["./main"]