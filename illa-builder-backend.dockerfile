# ---------------------
# build illa-builder-backend
FROM golang:1.20-bullseye as builder-for-backend

## set env
ENV LANG C.UTF-8
ENV LC_ALL C.UTF-8

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

## build
WORKDIR /opt/illa/illa-builder-backend
RUN cd  /opt/illa/illa-builder-backend
RUN ls -alh

COPY ./ ./

RUN cat ./Makefile

RUN make build-http-server

RUN ls -alh ./bin/illa-builder-backend


# -------------------
# build runner images
FROM alpine:latest as runner

WORKDIR /opt/illa/illa-builder-backend/bin/

## copy backend bin
COPY --from=builder-for-backend /opt/illa/illa-builder-backend/bin/illa-builder-backend /opt/illa/illa-builder-backend/bin/


RUN ls -alh /opt/illa/illa-builder-backend/bin/



# run
EXPOSE 8001
CMD ["/bin/sh", "-c", "/opt/illa/illa-builder-backend/bin/illa-builder-backend"]
