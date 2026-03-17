FROM node:20

RUN npm install -g @openai/codex

RUN mkdir -p /home/node/.codex
RUN chown -R node:node /home/node/.codex

USER node

WORKDIR /app
ENTRYPOINT ["bash"]
