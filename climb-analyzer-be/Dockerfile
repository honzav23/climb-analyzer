FROM golang:latest AS be

ENV PRODUCTION=true

WORKDIR /app/BE

COPY climb-analyzer-be/go.* ./
RUN go mod download

COPY climb-analyzer-be/*.go ./
COPY climb-analyzer-be/types ./types
COPY climb-analyzer-be/climbAnalyzer ./climbAnalyzer

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o ./climb-analyzer-server

RUN chmod +x ./climb-analyzer-server

EXPOSE 8080

# Run
CMD ["./climb-analyzer-server"]
