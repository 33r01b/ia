FROM node:20

RUN apt-get update \
    && apt-get install -y --no-install-recommends ripgrep procps \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g @openai/codex

RUN mkdir -p /home/node/.codex
RUN chown -R node:node /home/node/.codex

USER node

WORKDIR /app
CMD ["codex"]
