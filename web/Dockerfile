FROM node:16.15 as builder

RUN npm install -g pnpm@7.1.0

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN pnpm install

COPY . .

ARG VITE_RTC_ROOM_URL
ARG VITE_SFU_TURN_URLS
ARG VITE_SFU_STUN_URLS
ARG VITE_STORAGE_URL

ENV VITE_RTC_ROOM_URL=$VITE_RTC_ROOM_URL
ENV VITE_SFU_TURN_URLS=$VITE_SFU_TURN_URLS
ENV VITE_SFU_STUN_URLS=$VITE_SFU_STUN_URLS
ENV VITE_STORAGE_URL=$VITE_STORAGE_URL

RUN NODE_OPTIONS=--max-old-space-size=2048 pnpm run build

FROM nginx:1.19.6-alpine

COPY --from=builder /app/dist /etc/nginx/dist
