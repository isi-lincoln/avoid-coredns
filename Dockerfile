FROM golang:1.22-bullseye as build

RUN export DEBCONF_NONINTERACTIVE_SEEN=true \
           DEBIAN_FRONTEND=noninteractive \
           DEBIAN_PRIORITY=critical \
           TERM=linux ; \
    apt-get -qq update ; \
    apt-get -yyqq upgrade ; \
    apt-get -yyqq install \
        ca-certificates \
        libcap2-bin \
        build-essential \
        make; \
    apt-get clean

COPY . /coredns
WORKDIR /coredns
RUN ls
RUN make coredns
#RUN setcap cap_net_bind_service=+ep /coredns

FROM debian:stable-slim as final
RUN apt-get update && \
    apt-get install -qy \
        iputils-ping \
        iproute2 \
        dnsutils \
        vim-nox && \
    apt-get clean
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /coredns /coredns
EXPOSE 53 53/udp
EXPOSE 853 853/udp
CMD /coredns/coredns -conf /etc/coredns/Corefile
