FROM node:20

RUN npm install -g @google/gemini-cli

RUN mkdir -p /home/node/.gemini
RUN chown -R node:node /home/node/.gemini

USER node

WORKDIR /app
ENTRYPOINT ["bash"]
