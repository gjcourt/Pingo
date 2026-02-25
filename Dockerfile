FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /pingo ./cmd/ddns

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /pingo /usr/local/bin/pingo
USER 65534:65534
ENTRYPOINT ["pingo"]
