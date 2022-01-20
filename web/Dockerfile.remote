FROM node:14.16 as builder

RUN npm install -g pnpm

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN pnpm install

COPY . .

RUN pnpm run build

FROM nginx:1.19.6-alpine

COPY --from=builder /app/dist /etc/nginx/dist
