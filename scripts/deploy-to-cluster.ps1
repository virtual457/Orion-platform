# deploy-to-cluster.ps1
# Deploy Orion Platform to real Kubernetes cluster

Write-Host "Deploying Orion Platform to Kubernetes" -ForegroundColor Green
Write-Host "======================================="

# Check cluster access
Write-Host "`nStep 1: Verifying cluster access" -ForegroundColor Yellow
try {
    kubectl get nodes
    Write-Host "Cluster is accessible" -ForegroundColor Green
} catch {
    Write-Host "ERROR: Cannot access cluster" -ForegroundColor Red
    exit 1
}

Write-Host "`nStep 2: Building controller container" -ForegroundColor Yellow

# Create Dockerfile
$dockerfileContent = @'
FROM golang:1.21-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY pkg/ pkg/
RUN CGO_ENABLED=0 GOOS=linux go build -o controller ./cmd/operator

FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /workspace/controller .
USER 1001
ENTRYPOINT ["/controller"]
'@

$dockerfileContent | Out-File -FilePath "Dockerfile" -Encoding UTF8

# Build image
Write-Host "Building container image..." -ForegroundColor Cyan
docker build -t orion-platform:latest .

if ($LASTEXITCODE -eq 0) {
    Write-Host "Container built successfully" -ForegroundColor Green
} else {
    Write-Host "ERROR: Container build failed" -ForegroundColor Red
    exit 1
}

# Load into Kind
Write-Host "Loading image into cluster..." -ForegroundColor Cyan
kind load docker-image orion-platform:latest --name orion-platform

Write-Host "`nStep 3: Installing CRDs" -ForegroundColor Yellow
kubectl apply -f config\crd\application-crd.yaml

Write-Host "Waiting for CRDs..." -ForegroundColor Cyan
Start-Sleep -Seconds 5
kubectl get crd applications.platform.orion.dev

Write-Host "`nStep 4: Deploying controller" -ForegroundColor Yellow
kubectl apply -f deploy\controller.yaml

Write-Host "Waiting for controller..." -ForegroundColor Cyan
kubectl wait --for=condition=available --timeout=60s deployment/orion-controller -n orion-system

Write-Host "`nStep 5: Checking deployment" -ForegroundColor Yellow
kubectl get pods -n orion-system
kubectl get deployment -n orion-system

Write-Host "`nController logs:" -ForegroundColor Cyan
kubectl logs -n orion-system deployment/orion-controller --tail=5

Write-Host "`nSUCCESS! Controller deployed to real Kubernetes!" -ForegroundColor Green

Write-Host "`nNext: Create test application" -ForegroundColor Yellow
Write-Host "kubectl apply -f examples/test-app.yaml" -ForegroundColor Gray

# Clean up
Remove-Item Dockerfile -ErrorAction SilentlyContinue