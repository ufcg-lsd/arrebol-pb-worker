FROM golang:latest

WORKDIR /go/src/github.com/ufcg-lsd/arrebol-pb-worker/

RUN echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc

COPY ./core ./

WORKDIR /go/src/github.com/ufcg-lsd/arrebol-pb-worker/core

RUN go get -v
RUN go install -v

RUN go build -o main

CMD ('/bin/sh')
