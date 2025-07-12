FROM golang:1.24

WORKDIR app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /todoapp

RUN mkdir /db

EXPOSE 7540

ENV TODO_PASSWORD="666"

ENV TODO_DBFILE="/db/mydb.db"

VOLUME /db

CMD ["/todoapp"]