FROM oven/bun:1.1.21-debian@sha256:c59bbb19520a5d9417c0848496758ce0d20a6054500f0916b0939540df253839 AS bun

WORKDIR /app

COPY package.json bun.lockb tailwind.config.js tsconfig.json style.css ./

COPY ./internal/template/*.templ ./internal/template/

RUN bun install

RUN bunx tailwindcss -i ./style.css -o ./public/style.css



FROM golang:1.22-bullseye@sha256:583d5af8289d30de50aa0dcf4985d8b8746e52622becd6e1a62cfe191d5275a5 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download


COPY . .

COPY --from=bun /app/public/style.css /app/public/style.css

RUN go install github.com/a-h/templ/cmd/templ@v0.2.680

RUN templ generate

RUN go build \
  -ldflags="-linkmode external -extldflags -static -X 'main.BUILDTIME=$(date --iso-8601=seconds --utc)'" \
  -o api \
  ./cmd/api/main.go

RUN useradd -u 1001 crocoderdev


FROM scratch

WORKDIR /

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=build /etc/passwd /etc/passwd

COPY --from=build /app/api /api

COPY --from=build /app/public /public

COPY --from=build /app/internal/template/script /internal/template/script

COPY --from=build /app/internal/template/stylesheet /internal/template/stylesheet

COPY --from=build /app/internal/template/demo.html /internal/template/demo.html

USER crocoderdev

EXPOSE 80

EXPOSE 443

CMD ["/api"]
