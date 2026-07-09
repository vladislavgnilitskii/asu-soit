# --- сборка ---
FROM golang:1.26-alpine AS build
WORKDIR /src

# сначала зависимости — слой кэшируется, пока go.mod/go.sum не менялись
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# статический бинарь (CGO off) — чтобы работал на distroless без libc
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/server ./cmd/server

# --- рантайм ---
# distroless static: только бинарь, без shell и пакетов — минимум поверхности атаки
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/server"]
