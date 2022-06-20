FROM golang:1.18-alpine

ENV LABEL_TO_WATCH="fiware.cm-to-service"
ENV CREATED_LABEL_VALUE="k8s-cm-to-service"
ENV NAMESPACE_TO_WATCH=""

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./main.go ./main.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["k8s-cm-to-service"]