# syntax=docker/dockerfile:1.18

FROM golang:1.27.4-alpine3.23@sha256:f70850ef6ce341e8b6427f677049145d63c0b7ce002d4d7e3d464828b1b96dcd AS dependencies
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify

FROM dependencies AS build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly \
    -trimpath \
    -buildvcs=false \
    -ldflags="-s -w" \
    -o /out/stayforlong ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot@sha256:1b7b9f0f0e0a1d2155f531db587cc48ec26aaf97ab64364225f5bf18a054e66a AS runtime
WORKDIR /app
COPY --from=build --chown=nonroot:nonroot /out/stayforlong /app/stayforlong
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app/stayforlong", "healthcheck"]
ENTRYPOINT ["/app/stayforlong"]
