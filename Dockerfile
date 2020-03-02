FROM golang:1.14 as dev

ENV GO111MODULE=on

EXPOSE 8080

WORKDIR /unfail2ban

RUN go get github.com/go-task/task/v2/cmd/task \
    github.com/go-delve/delve/cmd/dlv

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . . 

RUN go install github.com/UCCNetsoc/UnFail2Ban

RUN go mod vendor

CMD [ "go", "run", "main.go" ]

FROM scratch

COPY --from=dev /go/bin/UnFail2Ban ./UnFail2Ban

CMD [ "./UnFail2Ban" ]