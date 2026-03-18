FROM debian:bookworm-slim

ARG ALL_PROXY
ARG NO_PROXY

ENV ALL_PROXY=$ALL_PROXY
ENV NO_PROXY=$NO_PROXY

# База (git/bash/curl) + типичные расширения (подстрой под проект)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates curl git bash unzip \
  && rm -rf /var/lib/apt/lists/*

# Создаём пользователя ДО установки
RUN useradd -m -u 1000 -s /bin/bash agent
USER agent

# Важно: HOME пользователя
ENV HOME=/home/agent
ENV DISABLE_AUTOUPDATER=1

# Часто launcher кладётся в ~/.local/bin
ENV PATH="/home/agent/.local/bin:${PATH}"

# Claude Code (native install, recommended)
RUN curl -fsSL https://claude.ai/install.sh | bash

WORKDIR /app
CMD ["claude"]
