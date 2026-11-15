# syntax=docker/dockerfile:1
FROM golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -mod=readonly -trimpath -ldflags="-s -w" -o /portfolio .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /portfolio /portfolio
ENV PORT=8080 APP_ENV=production
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/portfolio"]
