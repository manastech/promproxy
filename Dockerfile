FROM golang:1.12-alpine AS mods
RUN apk add --no-cache git
ADD go.mod /usr/src/promproxy/go.mod
ADD go.sum /usr/src/promproxy/go.sum
WORKDIR /usr/src/promproxy
RUN go mod download

FROM golang:1.12-alpine AS build
ADD . /usr/src/promproxy/
COPY --from=mods /go/pkg/mod/ /go/pkg/mod/
WORKDIR /usr/src/promproxy

RUN go install promproxy


FROM alpine
COPY --from=build /go/bin/promproxy /usr/bin/promproxy

CMD ["/usr/bin/promproxy"]
