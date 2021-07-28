FROM golang:1.16

WORKDIR /home/useradd/IITKCoin/Project
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["go","main/main.go"]