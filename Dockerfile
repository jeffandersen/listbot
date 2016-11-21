FROM alpine:latest

WORKDIR "/opt"

ADD .docker_build/repo-starbot /opt/bin/repo-starbot

CMD ["/opt/bin/repo-starbot"]
