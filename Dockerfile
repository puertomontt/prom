FROM golang:1.21.5

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -buildvcs=false -o ./main

CMD [ "./main" ]