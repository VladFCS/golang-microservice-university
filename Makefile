MODULE := github.com/vlad/microservices-grpc-kubernetes
PROTO_FILES := api/proto/catalog/v1/catalog.proto api/proto/inventory/v1/inventory.proto
GOBIN := $(shell go env GOPATH)/bin
IMAGE_TAG ?= dev
GHCR_NAMESPACE ?= ghcr.io/vladfcs
KYVERNO_VERSION ?= v1.15.2

.PHONY: proto tidy build test run-gateway run-catalog run-inventory docker-build minikube-up deploy unsigned-demo-push kyverno-install kyverno-policies-apply kyverno-demo-bad-latest kyverno-demo-bad-run-as-nonroot kyverno-demo-bad-privileged kyverno-demo-bad-hostnetwork kyverno-demo-bad-hostpath kyverno-demo-bad-no-resources kyverno-demo-unsigned kyverno-demo-signed demo-unsigned demo-latest demo-privileged demo-hostnetwork demo-no-limits demo-good argocd-install argocd-app-apply argocd-ui argocd-admin-password argocd-status

proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.10
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	PATH="$(GOBIN):$(PATH)" protoc \
		--proto_path=api/proto \
		--go_out=. \
		--go_opt=module=$(MODULE) \
		--go-grpc_out=. \
		--go-grpc_opt=module=$(MODULE) \
		$(PROTO_FILES)

tidy:
	go mod tidy

build:
	go build ./...

test:
	go test ./...

run-gateway:
	go run ./cmd/gateway-service

run-catalog:
	go run ./cmd/catalog-service

run-inventory:
	go run ./cmd/inventory-service

docker-build:
	docker build -f deployments/docker/gateway-service.Dockerfile -t $(GHCR_NAMESPACE)/gateway-service:$(IMAGE_TAG) .
	docker build -f deployments/docker/catalog-service.Dockerfile -t $(GHCR_NAMESPACE)/catalog-service:$(IMAGE_TAG) .
	docker build -f deployments/docker/inventory-service.Dockerfile -t $(GHCR_NAMESPACE)/inventory-service:$(IMAGE_TAG) .

minikube-up:
	minikube start --driver=docker --kubernetes-version=v1.30.0 --container-runtime=containerd
	kubectl config use-context minikube

deploy:
	kubectl apply -f deploy/k8s/namespace.yaml
	kubectl apply -k deploy/k8s

unsigned-demo-push:
	docker buildx build --platform linux/amd64,linux/arm64 --push -f deployments/docker/gateway-service.Dockerfile -t ghcr.io/vladfcs/golang-microservice-university-gateway-service:unsigned-demo .

kyverno-install:
	kubectl apply --server-side -f https://github.com/kyverno/kyverno/releases/download/$(KYVERNO_VERSION)/install.yaml

argocd-install:
	bash scripts/argocd_install.sh

argocd-app-apply:
	kubectl apply -f deploy/argocd/demo-application.yaml

argocd-ui:
	kubectl port-forward svc/argocd-server -n argocd 8088:443

argocd-admin-password:
	kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo

argocd-status:
	kubectl get pods -n argocd
	kubectl get applications -n argocd

kyverno-policies-apply:
	kubectl apply -k deploy/kyverno

kyverno-demo-bad-latest:
	kubectl apply -f deploy/kyverno/demo/bad-latest-deployment.yaml

kyverno-demo-bad-run-as-nonroot:
	kubectl apply -f deploy/kyverno/demo/bad-run-as-nonroot-deployment.yaml

kyverno-demo-bad-privileged:
	kubectl apply -f deploy/kyverno/demo/bad-privileged-deployment.yaml

kyverno-demo-bad-hostnetwork:
	kubectl apply -f deploy/kyverno/demo/bad-hostnetwork-deployment.yaml

kyverno-demo-bad-hostpath:
	kubectl apply -f deploy/kyverno/demo/bad-hostpath-deployment.yaml

kyverno-demo-bad-no-resources:
	kubectl apply -f deploy/kyverno/demo/bad-no-resources-deployment.yaml

kyverno-demo-unsigned:
	kubectl apply -f deploy/kyverno/demo/unsigned-image-deployment.yaml

kyverno-demo-signed:
	kubectl apply -f deploy/kyverno/demo/signed-image-deployment.yaml

demo-unsigned:
	kubectl apply -f demo/unsigned.yaml

demo-latest:
	kubectl apply -f demo/latest.yaml

demo-privileged:
	kubectl apply -f demo/privileged.yaml

demo-hostnetwork:
	kubectl apply -f demo/hostnetwork.yaml

demo-no-limits:
	kubectl apply -f demo/no-limits.yaml

demo-good:
	kubectl apply -f demo/good.yaml
