FROM golang:1.23 AS build

WORKDIR /rso-game

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./


RUN CGO_ENABLED=0 GOOS=linux go build -o /game_service

FROM gcr.io/distroless/base-debian11 AS release

WORKDIR /

COPY --from=build /game_service /game_service
COPY --from=build /rso-game/public /public

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT [ "/game_service" ]