# Start from golang base image
FROM golang:alpine as builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Set the current working directory inside the container 
WORKDIR /app

# Copy go mod and sum files 
COPY go.mod go.sum ./

# Download all dependencies.
RUN go mod download 

# Copy the source from the current directory to the working Directory inside the container 
COPY . .

# Run app
CMD ["go", "run", "main.go"]