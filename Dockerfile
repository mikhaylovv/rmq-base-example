FROM golang:1.27.2-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY publisher/ ./publisher/
COPY consumer/ ./consumer/
RUN CGO_ENABLED=0 go build -trimpath -o /out/publisher ./publisher && \
    CGO_ENABLED=0 go build -trimpath -o /out/consumer ./consumer

FROM scratch AS publisher
COPY --from=build /out/publisher /publisher
USER 65532:65532
ENTRYPOINT ["/publisher"]

FROM scratch AS consumer
COPY --from=build /out/consumer /consumer
USER 65532:65532
ENTRYPOINT ["/consumer"]
