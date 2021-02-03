FROM caddy:2.2.1-builder AS builder

COPY . /tmp/deviate-dns

RUN xcaddy build \
    --with github.com/rlweb/deviate-dns=/tmp/deviate-dns \
    --with github.com/abreka/caddy-tlsfirestore

FROM caddy:2.2.1

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
COPY caddy.json /etc/caddy/caddy.json

CMD ["caddy", "run", "--config", "/etc/caddy/caddy.json"]
