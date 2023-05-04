FROM golang@sha256:48c87cd759e3342fcbc4241533337141e7d8457ec33ab9660abe0a4346c30b60

RUN apk add make git \
    && mkdir -p /work
WORKDIR /work