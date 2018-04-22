# builder image
FROM golang:1.9-alpine as builder

RUN apk --no-cache add git
RUN go get github.com/golang/dep/cmd/dep
WORKDIR /go/src/github.com/paulmaddox/example-k8s-controller
COPY . .
RUN dep ensure
RUN go test -v ./...
RUN go build -o /bin/app -v \
  -ldflags "-X main.version=$(git describe --tags --always --dirty) -w -s"

# final image
FROM alpine:3.7
LABEL maintainer="Paul Maddox <pmaddox@amazon.com>"
EXPOSE 8080
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/app /bin/app
USER nobody
ENTRYPOINT ["/bin/app"]
