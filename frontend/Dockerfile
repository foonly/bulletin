FROM node:20-alpine

RUN npm install -g pnpm

WORKDIR /app

COPY pnpm-lock.yaml* package.json ./

RUN pnpm install

COPY . .

EXPOSE 5173

CMD ["pnpm", "run", "dev", "--", "--host"]
