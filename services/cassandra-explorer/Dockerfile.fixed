FROM node:18-alpine

WORKDIR /app

RUN npm install cassandra-driver express cors body-parser

COPY server.js .

EXPOSE 3000
CMD ["node", "server.js"]
