# syntax=docker/dockerfile:1

FROM golang:1.22.1

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./
COPY .env ./
# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /pikabu


# Run
CMD ["/pikabu"]