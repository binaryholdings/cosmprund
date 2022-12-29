# 1st stage, build app
FROM golang:1.19-alpine as builder

COPY . /app

WORKDIR /app

RUN go build -o cosmprund


FROM alpine

COPY --from=builder /app/cosmprund /usr/bin/cosmprund

ENTRYPOINT [ "/usr/bin/cosmprund" ]
