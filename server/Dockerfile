FROM golang:1.22-bookworm

ARG APT_MIRROR=mirrors.aliyun.com
RUN if [ -f /etc/apt/sources.list ]; then \
    sed -i "s/deb.debian.org/${APT_MIRROR}/g" /etc/apt/sources.list; \
    sed -i "s/security.debian.org/${APT_MIRROR}/g" /etc/apt/sources.list; \
  elif [ -f /etc/apt/sources.list.d/debian.sources ]; then \
    sed -i "s@http://deb.debian.org@http://${APT_MIRROR}@g" /etc/apt/sources.list.d/debian.sources; \
    sed -i "s@http://security.debian.org@http://${APT_MIRROR}@g" /etc/apt/sources.list.d/debian.sources; \
  fi \
  && apt-get update \
  && apt-get install -y --no-install-recommends ffmpeg ca-certificates git \
  && rm -rf /var/lib/apt/lists/*

ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=off
ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o server .

EXPOSE 18080

CMD ["./server"]
