FROM ubuntu:latest

ARG APP
ENV APP=$APP

ENV DEBIAN_FRONTEND noninteractive
RUN apt update && \
    apt install -y build-essential libpq-dev curl unzip

WORKDIR /
RUN curl https://github.com/vlang/v/releases/download/weekly.2022.46/v_linux.zip -L -O && \
  unzip v_linux.zip

WORKDIR /build
COPY . .

RUN mkdir -p /usr/local/$APP && /v/v -o /usr/local/$APP/app $APP

ENTRYPOINT stdbuf -i0 -o0 -e0 /usr/local/$APP/app