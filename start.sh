#!/bin/bash

# Build and start backend service
cd backend
go build -o app
./app &

# Build and start frontend service
cd ../frontend
npm install
npm run build
npm start 