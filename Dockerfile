# BUILD stage
FROM golang:1.13.8-alpine

# Install git for go download
RUN apk add git neofetch tzdata

# Set the timezone to europe
RUN cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime
RUN echo "Europe/Berlin" >  /etc/timezone
RUN date

# Copy files into the container and download dependencies
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Compile the server
RUN GO111MODULE=on CGO_ENABLED=0 go build

# Run the server
CMD ["./brobot"]
