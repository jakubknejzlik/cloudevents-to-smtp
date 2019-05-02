FROM golang as builder
WORKDIR /go/src/github.com/graphql-services/id
COPY . .
RUN go get ./... 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /tmp/app server/server.go

FROM jakubknejzlik/wait-for as wait-for

FROM alpine:3.5

WORKDIR /app

COPY --from=wait-for /usr/local/bin/wait-for /usr/local/bin/wait-for
COPY --from=builder /tmp/app /usr/local/bin/app
COPY --from=builder /go/src/github.com/graphql-services/id/schema.graphql /app/schema.graphql

# RUN apk --update add docker

# https://serverfault.com/questions/772227/chmod-not-working-correctly-in-docker
RUN chmod +x /usr/local/bin/app

ENTRYPOINT []
CMD [ "/bin/sh", "-c", "wait-for ${DATABASE_URL} && app server" ]