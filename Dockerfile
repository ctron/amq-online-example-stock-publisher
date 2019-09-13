FROM centos:7

RUN yum install -y git
RUN curl -L https://dl.google.com/go/go1.12.7.linux-amd64.tar.gz -o go.tar.gz 
RUN tar xzf go.tar.gz

COPY go.mod go.sum main.go /
RUN /go/bin/go build .

ENTRYPOINT /stock-publisher
