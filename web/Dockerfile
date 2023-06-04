FROM node:20.2 as builder

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

ARG VITE_RTC_ROOM_URL
ARG VITE_SFU_URL
ARG VITE_STORAGE_URL
ARG VITE_GSI_CLIENT_ID

RUN NODE_OPTIONS=--max-old-space-size=2048 npm run build

FROM nginx:1.19.6-alpine

COPY --from=builder /app/dist /etc/nginx/dist
