FROM golang:latest
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o fragmenter-sidecar .

CMD ["/app/fragmenter-sidecar"]