FROM node:14.16 as builder

RUN npm install -g pnpm

WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN pnpm install

COPY . .

ARG VITE_RTC_ROOM_URL
ARG VITE_RTC_TURN_URLS

ENV VITE_RTC_ROOM_URL=$VITE_RTC_ROOM_URL
ENV VITE_RTC_TURN_URLS=$VITE_RTC_TURN_URLS

RUN pnpm run build

FROM nginx:1.19.6-alpine

COPY --from=builder /app/dist /etc/nginx/dist
