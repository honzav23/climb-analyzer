FROM golang:latest AS be

ENV PRODUCTION=true

WORKDIR /app/BE

COPY climb-analyzer-be/go.* ./
RUN go mod download

COPY climb-analyzer-be/*.go ./

WORKDIR /app/BE/types

COPY climb-analyzer-be/types .

WORKDIR /app/BE

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o ./climbAnalyzer

RUN chmod +x ./climbAnalyzer

EXPOSE 8080

# Run
CMD ["./climbAnalyzer"]