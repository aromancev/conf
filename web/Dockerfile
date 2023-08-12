FROM node:20.2 as builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

ARG VITE_API_GRAPH_URL
ARG VITE_API_IAM_URL
ARG VITE_API_ROOM_RTC_URL
ARG VITE_SFU_URL
ARG VITE_STORAGE_URL
ARG VITE_GSI_CLIENT_ID

RUN npm run build

FROM nginx:1.19.6-alpine

COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /etc/nginx/dist
