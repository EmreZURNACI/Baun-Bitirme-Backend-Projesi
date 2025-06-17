FROM golang

RUN mkdir /app

WORKDIR /app

COPY . /app

RUN go mod tidy
RUN go mod download

CMD ["go","run","main.go"]

EXPOSE 8085