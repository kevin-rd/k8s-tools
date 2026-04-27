FROM bitnamilegacy/postgresql:16.4.0-debian-12-r9

USER root

RUN install_packages build-essential git postgresql-server-dev-16 ca-certificates

RUN cd /tmp \
    && git clone --branch v0.8.2 https://github.com/pgvector/pgvector.git \
    && cd pgvector \
    && make \
    && make install \
    && cp /usr/share/postgresql/16/extension/vector* /opt/bitnami/postgresql/share/extension/ \
    && cp /usr/lib/postgresql/16/lib/vector.so /opt/bitnami/postgresql/lib/ \
    && rm -rf /tmp/pgvector

USER 1001