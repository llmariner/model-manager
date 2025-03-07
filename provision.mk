LLMA_REPO ?= https://github.com/llmariner/llmariner.git
CLONE_PATH ?= work

.PHONY: reapply-model-server
reapply-model-server: load-server-image helm-apply-llma rollout-model-server
.PHONY: reapply-model-loader
reapply-model-loader: load-loader-image helm-apply-llma rollout-model-loader

# ------------------------------------------------------------------------------
# chart repository
# ------------------------------------------------------------------------------

.PHONY: pull-llma-chart
pull-llma-chart:
	@if [ -d $(CLONE_PATH) ]; then \
		cd $(CLONE_PATH) && \
		git checkout -- deployments/llmariner/Chart.yaml && \
		git pull; \
	else \
		git clone $(LLMA_REPO) $(CLONE_PATH); \
	fi

.PHONY: configure-llma-chart
configure-llma-chart:
	hack/overwrite-llma-chart-for-test.sh $(CLONE_PATH)
	-rm $(CLONE_PATH)/deployments/llmariner/Chart.lock

# ------------------------------------------------------------------------------
# deploy dependencies
# ------------------------------------------------------------------------------

DEP_APPS ?= minio,postgres

.PHONY: helm-apply-deps
helm-apply-deps:
	hack/helm-apply-deps.sh $(CLONE_PATH) $(DEP_APPS) kind-$(KIND_CLUSTER)

# ------------------------------------------------------------------------------
# deploy llmariner
# ------------------------------------------------------------------------------

KIND_CLUSTER ?= kind
EXTRA_VALS ?=

.PHONY: helm-apply-llma
helm-apply-llma:
	hack/helm-apply-llma.sh $(CLONE_PATH)

# ------------------------------------------------------------------------------
# load images
# ------------------------------------------------------------------------------

.PHONY: load-server-image
load-server-image: build-docker-server
	@kind load docker-image $(SERVER_IMAGE):$(TAG) --name $(KIND_CLUSTER)

.PHONY: load-loader-image
load-loader-image: build-docker-loader
	@kind load docker-image $(LOADER_IMAGE):$(TAG) --name $(KIND_CLUSTER)

# ------------------------------------------------------------------------------
# rollout pods
# ------------------------------------------------------------------------------

.PHONY: rollout-model-server
rollout-model-server:
	@kubectl --context kind-$(KIND_CLUSTER) rollout restart deployment -n llmariner model-manager-server

.PHONY: rollout-model-loader
rollout-model-loader:
	@kubectl --context kind-$(KIND_CLUSTER) rollout restart deployment -n llmariner model-manager-loader
