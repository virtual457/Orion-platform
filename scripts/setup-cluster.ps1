# ORION PLATFORM - CLEAN SETUP SCRIPT
# Simple, working script to install everything

#Requires -RunAsAdministrator

Write-Host "=========================================================" -ForegroundColor Green
Write-Host "ORION PLATFORM - COMPLETE SETUP" -ForegroundColor Green
Write-Host "=========================================================" -ForegroundColor Green

# Helper function
function Test-Command($cmdname) {
    return [bool](Get-Command -Name $cmdname -ErrorAction SilentlyContinue)
}

Write-Host "`nStep 1: Checking Prerequisites" -ForegroundColor Yellow

# Check Go
if (Test-Command "go") {
    $goVersion = go version
    Write-Host "Go: $goVersion" -ForegroundColor Green
} else {
    Write-Host "ERROR: Go not installed" -ForegroundColor Red
    exit 1
}

Write-Host "`nStep 2: Installing Docker Desktop" -ForegroundColor Yellow

# Check Docker
if (Test-Command "docker") {
    Write-Host "Docker: Already installed" -ForegroundColor Green
    
    # Test Docker daemon
    try {
        docker ps | Out-Null
        Write-Host "Docker daemon: Running" -ForegroundColor Green
    } catch {
        Write-Host "Docker daemon: Not running" -ForegroundColor Yellow
        
        # Start Docker Desktop
        $dockerPath = "${env:ProgramFiles}\Docker\Docker\Docker Desktop.exe"
        if (Test-Path $dockerPath) {
            Write-Host "Starting Docker Desktop..." -ForegroundColor Cyan
            Start-Process $dockerPath
            
            # Wait for Docker
            Write-Host "Waiting for Docker daemon..." -ForegroundColor Cyan
            $attempts = 0
            do {
                Start-Sleep -Seconds 10
                $attempts++
                try {
                    docker ps | Out-Null
                    Write-Host "Docker is ready!" -ForegroundColor Green
                    break
                } catch {
                    Write-Host "  Waiting... attempt $attempts" -ForegroundColor Gray
                }
            } while ($attempts -lt 15)
            
            # Final check
            try {
                docker ps | Out-Null
            } catch {
                Write-Host "ERROR: Docker failed to start" -ForegroundColor Red
                exit 1
            }
        }
    }
} else {
    Write-Host "Docker: Not installed - installing now..." -ForegroundColor Cyan
    
    # Try winget installation
    if (Test-Command "winget") {
        Write-Host "Installing via winget..." -ForegroundColor Cyan
        winget install Docker.DockerDesktop --accept-source-agreements --accept-package-agreements --silent
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Docker Desktop installed!" -ForegroundColor Green
            Write-Host "RESTART REQUIRED - Please restart computer and re-run script" -ForegroundColor Yellow
            exit 2
        }
    }
    
    # Try chocolatey if winget failed
    if (Test-Command "choco") {
        Write-Host "Installing via chocolatey..." -ForegroundColor Cyan
        choco install docker-desktop -y
        Write-Host "RESTART REQUIRED - Please restart computer and re-run script" -ForegroundColor Yellow
        exit 2
    }
    
    # Manual download if package managers not available
    Write-Host "Package managers not available - downloading manually..." -ForegroundColor Yellow
    
    $dockerUrl = "https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe"
    $installerPath = "$env:TEMP\DockerInstaller.exe"
    
    try {
        Write-Host "Downloading Docker Desktop..." -ForegroundColor Cyan
        Invoke-WebRequest -Uri $dockerUrl -OutFile $installerPath -UseBasicParsing
        
        Write-Host "Running installer..." -ForegroundColor Cyan
        Start-Process $installerPath -ArgumentList "install --quiet" -Wait
        
        Write-Host "Docker Desktop installed!" -ForegroundColor Green
        Write-Host "RESTART REQUIRED - Please restart computer and re-run script" -ForegroundColor Yellow
        
        Remove-Item $installerPath -ErrorAction SilentlyContinue
        exit 2
        
    } catch {
        Write-Host "ERROR: Failed to install Docker Desktop" -ForegroundColor Red
        Write-Host "Please install manually: https://www.docker.com/products/docker-desktop/" -ForegroundColor Yellow
        exit 1
    }
}

Write-Host "`nStep 3: Installing Kubernetes Tools" -ForegroundColor Yellow

# Install Kind
if (Test-Command "kind") {
    Write-Host "Kind: Already installed" -ForegroundColor Green
} else {
    Write-Host "Installing Kind..." -ForegroundColor Cyan
    go install sigs.k8s.io/kind@v0.20.0
    
    # Add to PATH
    $goPath = go env GOPATH
    if ($goPath) {
        $env:PATH = "$goPath\bin;$env:PATH"
    }
    
    if (Test-Command "kind") {
        Write-Host "Kind: Installed" -ForegroundColor Green
    } else {
        Write-Host "ERROR: Kind installation failed" -ForegroundColor Red
        exit 1
    }
}

# Install kubectl
if (Test-Command "kubectl") {
    Write-Host "kubectl: Already installed" -ForegroundColor Green
} else {
    Write-Host "Installing kubectl..." -ForegroundColor Cyan
    
    $goPath = go env GOPATH
    if (-not $goPath) {
        $goPath = "$env:USERPROFILE\go"
    }
    $binPath = "$goPath\bin"
    
    New-Item -ItemType Directory -Force -Path $binPath | Out-Null
    
    $kubectlUrl = "https://dl.k8s.io/release/v1.28.0/bin/windows/amd64/kubectl.exe"
    $kubectlPath = "$binPath\kubectl.exe"
    
    Invoke-WebRequest -Uri $kubectlUrl -OutFile $kubectlPath -UseBasicParsing
    $env:PATH = "$binPath;$env:PATH"
    
    if (Test-Command "kubectl") {
        Write-Host "kubectl: Installed" -ForegroundColor Green
    } else {
        Write-Host "ERROR: kubectl installation failed" -ForegroundColor Red
        exit 1
    }
}

Write-Host "`nStep 4: Creating Kubernetes Cluster" -ForegroundColor Yellow

# Create cluster config
$config = @'
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: orion-platform
nodes:
- role: control-plane
  image: kindest/node:v1.27.3
  extraPortMappings:
  - containerPort: 80
    hostPort: 8080
    protocol: TCP
  - containerPort: 9001
    hostPort: 9001
    protocol: TCP
  - containerPort: 5432
    hostPort: 5432
    protocol: TCP
'@

$config | Out-File -FilePath "cluster-config.yaml" -Encoding UTF8

# Delete existing cluster
kind delete cluster --name orion-platform 2>$null

# Create cluster
Write-Host "Creating Kubernetes cluster..." -ForegroundColor Cyan
kind create cluster --config cluster-config.yaml --wait 300s

if ($LASTEXITCODE -eq 0) {
    Write-Host "Cluster created!" -ForegroundColor Green
} else {
    Write-Host "ERROR: Cluster creation failed" -ForegroundColor Red
    exit 1
}

Write-Host "`nStep 5: Configuring Cluster" -ForegroundColor Yellow

# Set context
kubectl config use-context kind-orion-platform

# Create namespaces
kubectl create namespace orion-system --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace applications --dry-run=client -o yaml | kubectl apply -f -

Write-Host "`nStep 6: Building Platform" -ForegroundColor Yellow

# Build operator
go mod tidy
go build -o bin\operator.exe .\cmd\operator

if ($LASTEXITCODE -eq 0) {
    Write-Host "Operator built successfully!" -ForegroundColor Green
} else {
    Write-Host "ERROR: Build failed" -ForegroundColor Red
    exit 1
}

# Clean up
Remove-Item "cluster-config.yaml" -ErrorAction SilentlyContinue

Write-Host "`n=========================================================" -ForegroundColor Green
Write-Host "SUCCESS! ORION PLATFORM READY!" -ForegroundColor Green
Write-Host "=========================================================" -ForegroundColor Green

Write-Host "`nCluster: orion-platform" -ForegroundColor White
Write-Host "Context: kind-orion-platform" -ForegroundColor White

kubectl get nodes
kubectl get namespaces

Write-Host "`nNext: Deploy controller to real cluster!" -ForegroundColor Yellow
Write-Host "Command: kubectl apply -f deploy/" -ForegroundColor Gray

Write-Host "`nTest platform: .\bin\operator.exe" -ForegroundColor Yellow