FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY . .

ARG REACT_APP_BACKEND_HOST
ENV REACT_APP_BACKEND_HOST=$REACT_APP_BACKEND_HOST

RUN npm run build
RUN npm install -g serve

ENV NODE_ENV=production
ENV PORT=3000

EXPOSE 3000

CMD ["serve", "-s", "build", "-l", "3000"] 