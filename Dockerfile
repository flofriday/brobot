# BUILD stage
FROM golang:1.13.5-alpine

# Install git for go download
RUN apk add git neofetch

# Copy files into the container and download dependencies
#WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Compile the server
RUN GO111MODULE=on CGO_ENABLED=0 go build

# Run the server
CMD ["./brobot"]
