FROM golang
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/NewPipeLine
ADD /. .
RUN go install .

FROM alpine:latest
WORKDIR /root/
COPY --from=0 /go/bin/pipeline .
ENTRYPOINT ./NewPipeLine
EXPOSE 8080