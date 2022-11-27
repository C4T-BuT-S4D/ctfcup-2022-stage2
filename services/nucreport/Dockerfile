FROM rust:1.65 as builder
WORKDIR /usr/src/app
COPY app .

RUN openssl rand -hex 64 > /tmp/secret_key

RUN cargo install --path .

FROM debian:bullseye

RUN apt-get update && apt-get install -y openssh-server sudo && rm -rf /var/lib/apt/lists/*

RUN useradd --no-log-init -ms /usr/sbin/nologin nuclear
RUN sudo usermod -aG shadow nuclear

COPY conf/sudoers /etc/sudoers
COPY conf/adduser.conf /etc/adduser.conf
COPY conf/sshd_config /etc/ssh/sshd_config
COPY conf/newuser.sh /usr/sbin/addnewuser.sh
COPY conf/readfile.sh /usr/sbin/readfile.sh

RUN chmod +x /usr/sbin/addnewuser.sh && chmod +x /usr/sbin/readfile.sh

COPY entry.sh .
RUN chmod +x entry.sh

COPY --from=builder /usr/local/cargo/bin/app /usr/local/bin/app
COPY conf/app/.env .


CMD ["./entry.sh"]