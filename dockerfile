FROM golang:1.19 as build
WORKDIR /tmp/
COPY durin.go .
RUN go build durin.go && cp durin /durin

FROM alpine:latest
WORKDIR /app
RUN apk add gcompat
COPY --from=build /durin ./durin
ENTRYPOINT ["/app/durin", "-d", "/app/db", "-b", "0.0.0.0"]
