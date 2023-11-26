# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)), /bin/ash, /bin/bash)

# OPA Playground
# 	https://play.openpolicyagent.org/
# 	https://academy.styra.com/
# 	https://www.openpolicyagent.org/docs/latest/policy-reference/

# ==============================================================================
# Install dependencies

dev-gotooling:
	go install github.com/divan/expvarmon@latest
	go install github.com/rakyll/hey@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/goimports@latest

dev-brew:
	brew update
	brew tap hashicorp/tap
	brew list kind || brew install kind
	brew list kubectl || brew install kubectl
	brew list kustomize || brew install kustomize
	brew list pgcli || brew install pgcli
	brew list vault || brew install vault

dev-docker:
	docker pull $(GOLANG)
	docker pull $(ALPINE)
	docker pull $(KIND)
	docker pull $(POSTGRES)
	docker pull $(VAULT)
	docker pull $(GRAFANA)
	docker pull $(PROMETHEUS)
	docker pull $(TEMPO)
	docker pull $(LOKI)
	docker pull $(PROMTAIL)
	docker pull $(TELEPRESENCE)

# ==============================================================================
# Define dependencies

GOLANG          := golang:1.21.3
ALPINE          := alpine:3.18
KIND            := kindest/node:v1.27.3
POSTGRES        := postgres:15.4
VAULT           := hashicorp/vault:1.15
GRAFANA         := grafana/grafana:10.1.0
PROMETHEUS      := prom/prometheus:v2.47.0
TEMPO           := grafana/tempo:2.2.0
LOKI            := grafana/loki:2.9.0
PROMTAIL        := grafana/promtail:2.9.0
TELEPRESENCE    := datawire/tel2:2.13.2

KIND_CLUSTER    := ardan-starter-cluster
NAMESPACE       := sales-system
APP             := sales 
BASE_IMAGE_NAME := ardanlabs/service
SERVICE_NAME    := sales-api
VERSION         := 0.0.1
SERVICE_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)
METRICS_IMAGE   := $(BASE_IMAGE_NAME)/$(SERVICE_NAME)-metrics:$(VERSION)

# ==============================================================================
tidy:
		go mod tidy
		go mod vendor

run-local:
		go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go -service=$(SERVICE_NAME)

run-local-help:
		go run app/services/sales-api/main.go -h


metrics-view:
	expvarmon -ports="$(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

metrics-view-local:
	expvarmon -ports="localhost:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.Alloc"

# ==============================================================================
dev-up-local:
		kind create cluster \
			--image $(KIND) \
			--name $(KIND_CLUSTER) \
			--config zarf/k8s/dev/kind-config.yaml

		kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner
		kind load docker-image $(TELEPRESENCE) --name $(KIND_CLUSTER)
		kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)
		

# Helm is responsible for starting the telepresence service inside the cluser
# The second command is to start the agent service behind the scenes and create that tunnel for us
dev-up: dev-up-local dev-telepresence

dev-telepresence:
		
		telepresence --context=kind-$(KIND_CLUSTER) helm install
		telepresence --context=kind-$(KIND_CLUSTER) connect

dev-down-local:
		kind delete cluster --name $(KIND_CLUSTER)

dev-down:
		kind delete cluster --name $(KIND_CLUSTER)

dev-load:
	kind load docker-image $(SERVICE_IMAGE) --name $(KIND_CLUSTER)

dev-apply:
	kustomize build zarf/k8s/dev/database | kubectl apply -f -
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/database

	kustomize build zarf/k8s/dev/sales | kubectl apply -f -
	kubectl wait pods --namespace=$(NAMESPACE) --selector app=$(APP) --timeout=120s --for=condition=Ready

dev-restart:
	kubectl rollout restart deployment $(APP) --namespace=$(NAMESPACE)

# Here we push code changes and we are restarting the whole thing
dev-update: all dev-load dev-restart

# Here we push new configurations
dev-update-apply: all dev-load dev-apply

# ------------------------------------------------------------------------------

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -l app=$(APP) --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go -service=$(SERVICE_NAME)

dev-describe-deployment:
	kubectl describe deployment --namespace=$(NAMESPACE) $(APP)

dev-describe-sales:
	kubectl describe pod --namespace=$(NAMESPACE) -l app=$(APP)	

# ------------------------------------------------------------------------------
dev-status:
		kubectl get nodes -o wide
		kubectl get svc -o wide
		kubectl get pods -o wide --watch --all-namespaces

# ==============================================================================
# Building containers

all: service

service:
	docker build \
		-f zarf/docker/dockerfile.service \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ==============================================================================
# Curling the service
test-endpoint:
	curl -il http://$(SERVICE_NAME).$(NAMESPACE).svc.cluster.local:4000/debug/pprof

test-endpoint-local:
	curl -il http://localhost:4000/debug/pprof

#===============================================================================
# RSA Keys
# 	To generate a private/public key PEM file.
# 	$ openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# 	$ openssl rsa -pubout -in private.pem -out public.pem

run-scratch:
	go run app/scratch/main.go

# ==============================================================================
# Administration
pgcli-local:
	pgcli postgresql://postgres:postgres@localhost

pgcli:
	pgcli postgresql://postgres:postgres@database-service.${NAMESPACE}.svc.cluser.local