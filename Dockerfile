FROM golang@sha256:48c87cd759e3342fcbc4241533337141e7d8457ec33ab9660abe0a4346c30b60

RUN apk add make git \
    && mkdir -p /work
COPY . /work
WORKDIR /work
RUN make build

FROM 494198612820.dkr.ecr.us-west-2.amazonaws.com/shared/appdynamics-hybridagent-jre:17-alpine-hybrid-agent-latest
COPY --from=0 /work/goofys /usr/bin/goofys
RUN chmod +x /usr/bin/goofys
RUN apk add --no-cache curl fuse \
    && mkdir -p /opt/appdynamics \
    && chown -R appdynamics:appdynamics /opt/appdynamics \
    && curl -L https://artifactory.bare.appdynamics.com/artifactory/github/open-telemetry/opentelemetry-java-instrumentation/releases/download/v1.17.0/opentelemetry-javaagent.jar > /usr/local/lib/opentelemetry-javaagent.jar \
    && chown -R appdynamics:appdynamics /usr/local/lib/opentelemetry-javaagent.jar \
    && mkdir /mnt/goofys \
    && chown -R appdynamics:appdynamics /mnt/goofys \
    && mkdir -p /licenses \
    && chown -R appdynamics:appdynamics /licenses
COPY ./base/licenses/LICENSE /licenses
