# Makefile for starting the Social Network app

.PHONY: all backend frontend start-backend start-frontend

# Run both backend and frontend
all: start-backend start-frontend

# Start the backend
start-backend:
	cd "backend" && go run .

# Install dependencies and start the frontend
start-frontend:
	cd "client" && npm install && npm run dev
