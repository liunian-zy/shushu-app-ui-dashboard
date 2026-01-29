FROM golang:1.22-bookworm

RUN apt-get update \
  && apt-get install -y --no-install-recommends ffmpeg ca-certificates git \
  && rm -rf /var/lib/apt/lists/*

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o server .

EXPOSE 18080

CMD ["./server"]
