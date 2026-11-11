ARG GO_VERSION=1.27
FROM golang:${GO_VERSION}-bookworm

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
	&& apt-get install -y --no-install-recommends \
		make \
		git \
		ca-certificates \
		gcc \
		curl \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /workspaces/web-analyzer
