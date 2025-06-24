FROM docker.io/library/golang:1.23 as build

WORKDIR /src

COPY ./.. .

RUN go mod tidy
RUN go build -o app cmd/app/main.go


FROM docker.io/library/golang:1.23

WORKDIR /src

COPY --from=build /src/app app
COPY ./cmd/app/config.yml ./cmd/app/config.yml
COPY ./internal/repository/part1.txt ./internal/repository/part1.txt
COPY ./internal/repository/part2.txt ./internal/repository/part2.txt

CMD ["./app"]
