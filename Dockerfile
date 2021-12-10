FROM golang:alpine

RUN apk update

RUN apk add git

RUN apk add make

RUN apk add perl

RUN git clone https://github.com/darold/pgFormatter.git /pgformatter

WORKDIR /pgformatter

RUN perl Makefile.PL

RUN make && make install

WORKDIR /goapp

COPY . /goapp

RUN go build -o /sqlfmt

ENTRYPOINT ["/sqlfmt"]
