# start-app.ps1
<#
    Script to streamline starting the Social Network app
#>

# Function to run the backend
function Start-Backend {
    Write-Host "Starting Backend..." -ForegroundColor Cyan
    Push-Location "backend"
    go run .
    Pop-Location
}

# Function to run the frontend
function Start-Frontend {
    Write-Host "Starting Frontend..." -ForegroundColor Cyan
    Push-Location "client"
    npm install
    npm run dev
    Pop-Location
}

# Main function
Write-Host "Starting Social Network App..." -ForegroundColor Green
Start-Backend
Start-Frontend
