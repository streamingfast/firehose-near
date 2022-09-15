FROM ubuntu:20.04

RUN DEBIAN_FRONTEND=noninteractive apt-get update && \
       apt-get -y install -y \
       ca-certificates libssl1.1 vim htop iotop sysstat \
       dstat strace lsof curl jq tzdata && \
       rm -rf /var/cache/apt /var/lib/apt/lists/*

RUN rm /etc/localtime && ln -snf /usr/share/zoneinfo/America/Montreal /etc/localtime && dpkg-reconfigure -f noninteractive tzdata

ADD /firenear /app/firenear

COPY tools/firenear/motd_generic /etc/motd
COPY tools/firenear/99-firenear.sh /etc/profile.d/

# On SSH connection, /root/.bashrc is invoked which invokes '/root/.bash_aliases' if existing,
# so we hijack the file to "execute" our specialized bash script
RUN echo ". /etc/profile.d/99-firenear.sh" > /root/.bash_aliases

ENV PATH "$PATH:/app"

ENTRYPOINT ["/app/firenear"]
