FROM golang:1.16.5-alpine3.14 as BUILD

WORKDIR /usr/src/app

COPY . ${WORKDIR}

RUN apk update && apk add bash
RUN go get ./...
RUN go build -o main src/github.com/bludot/goweather/*.go


FROM golang:1.16.5-alpine3.14 as FINAL

WORKDIR /usr/src/app

RUN touch .env
COPY --from=BUILD /usr/src/app/main ./main

CMD ["./main"]
