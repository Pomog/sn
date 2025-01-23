<#
    Script to streamline starting the Social Network app
#>

# Function to run the backend in a new terminal window
function Start-Backend {
    Write-Host "Starting Backend..." -ForegroundColor Cyan
    Start-Process "powershell.exe" -ArgumentList "-NoExit", "-Command", "cd backend; go run ."
}

# Function to run the frontend in a new terminal window
function Start-Frontend {
    Write-Host "Starting Frontend..." -ForegroundColor Cyan
    Start-Process "powershell.exe" -ArgumentList "-NoExit", "-Command", "cd client; npm install; npm run dev"
}

# Main function
Write-Host "Starting Social Network App..." -ForegroundColor Green
Start-Backend
Start-Frontend
