FROM node:16.15 as builder

RUN npm install -g pnpm@7.1.0

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN pnpm install

COPY . .

ARG VITE_RTC_ROOM_URL
ARG VITE_SFU_URL
ARG VITE_STORAGE_URL
ARG VITE_GSI_CLIENT_ID

RUN NODE_OPTIONS=--max-old-space-size=2048 pnpm run build

FROM nginx:1.19.6-alpine

COPY --from=builder /app/dist /etc/nginx/dist
