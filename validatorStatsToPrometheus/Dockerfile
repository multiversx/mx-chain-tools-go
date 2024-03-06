FROM golang:1.20.7 as builder
MAINTAINER MultiversX


WORKDIR /app
COPY . .

WORKDIR /app/validatorStatisticsToPrometheus

RUN go build


FROM ubuntu:22.04
RUN apt-get update && apt-get install openssl -y

WORKDIR /app
RUN useradd -s /bin/bash -u 1001 ubuntu
RUN chown ubuntu /app

COPY --from=builder --chown=ubuntu:ubuntu  "/app/validatorStatisticsToPrometheus/validatorStatisticsToPrometheus" "/app/validatorStatisticsToPrometheus"
COPY --from=builder --chown=ubuntu:ubuntu  "/app/validatorStatisticsToPrometheus/config.toml" "/app/config.toml"
COPY --chown=ubuntu:ubuntu  ./scripts/entrypoint.sh "/app/entrypoint.sh"

RUN chmod +x /app/validatorStatisticsToPrometheus

USER ubuntu
EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
