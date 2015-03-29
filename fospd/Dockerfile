FROM google/golang
MAINTAINER felix.maurer@student.kit.edu


RUN go get github.com/maufl/go-fosp/fospd

ADD config.json /etc/fospd/config.json

EXPOSE 1337
EXPOSE 1338

ENTRYPOINT ["/gopath/bin/fospd"]
CMD []
