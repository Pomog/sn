#!/bin/bash

# Script to streamline starting the Social Network app

# Function to run the backend in a new terminal window
start_backend() {
    echo "Starting Backend..."
    gnome-terminal -- bash -c "cd backend; go run .; exec bash"
}

# Function to run the frontend in a new terminal window
start_frontend() {
    echo "Starting Frontend..."
    gnome-terminal -- bash -c "cd client; npm install; npm run dev; exec bash"
}

# Main function
echo "Starting Social Network App..."
start_backend
start_frontend
