# Étape 1 : build
FROM node:20-alpine AS builder

WORKDIR /app

# Dépendances
COPY package*.json ./
RUN npm install

# Prisma
COPY prisma ./prisma
RUN npx prisma generate

# Code source
COPY . .
RUN npm run build

# Étape 2 : runtime
FROM node:20-alpine

WORKDIR /app

ENV NODE_ENV=production

COPY package*.json ./
RUN npm install --omit=dev

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/prisma ./prisma

EXPOSE 3003

CMD ["node", "dist/main.js"]
