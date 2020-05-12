# Build
FROM golang:1.14-alpine AS builder

RUN apk add --no-cache gcc musl-dev linux-headers

COPY . /quorum-reporting
RUN cd /quorum-reporting && go build -o reporting

# Deployment
FROM alpine:latest

COPY --from=builder /quorum-reporting/reporting /usr/local/bin/

ENTRYPOINT ["reporting"]