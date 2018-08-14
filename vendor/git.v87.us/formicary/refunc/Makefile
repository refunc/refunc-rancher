SHELL := /bin/bash # ensure bash is used

BINS := agent bundle credsyncer

images: $(addsuffix -image, $(BINS))

$(BINS): % : bin/%
	@source hack/scripts/common; log_info "Building $@ Done"

bin/%:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build \
	-tags netgo -installsuffix netgo \
	-ldflags "-s -w $(LD_FLAGS)" \
	-a \
	-o $@ \
	./cmd/$*/

%-image: % package/Dockerfile
	@rm package/$* 2>/dev/null || true && cp bin/$* package/
	@ source hack/scripts/common \
	&& cd package \
	&& docker build \
	--build-arg https_proxy="${HTTPS_RPOXY}" \
	--build-arg http_proxy="${HTTP_RPOXY}" \
	--build-arg BIN_TARGET=$* \
	-t $(IMAGE) .
	@source hack/scripts/common; \
	if ! echo $(IMAGE) | grep -qs 'dirty'; then \
		log_info pushing $(IMAGE); \
	else \
		log_warn pushing dirty images $(IMAGE); \
	fi
	@docker push $(IMAGE);

AGENT_IMAGE=$(shell source hack/scripts/version; echo $${AGENT_IMAGE})
bin/agent: cmd/agent/*.go pkg/runtime/agent/loader/*.go
agent-image: IMAGE=$(AGENT_IMAGE)

BUNDLE_VERSION=$(shell source hack/scripts/version; echo $${BUNDLE_VERSION})
BUNDLE_IMAGE=$(shell source hack/scripts/version; echo $${BUNDLE_IMAGE})
bin/bundle: LD_FLAGS=-X git.v87.us/formicary/refunc/pkg/runtime/agent/runtime.InitContainerImage=$(AGENT_IMAGE) \
-X main.Version=$(BUNDLE_VERSION) \
-X git.v87.us/formicary/refunc/pkg/buildins/helloworld.BundleVersion=$(BUNDLE_VERSION) \
-X git.v87.us/formicary/refunc/pkg/buildins/helloworld.AgentVersion=$(shell source hack/scripts/version; echo $${AGENT_VERSION})
bin/bundle: $(shell find pkg -type f -name '*.go') $(shell find cmd -type f -name '*.go')
bundle-image: IMAGE=$(BUNDLE_IMAGE)
bundle-image: agent-image

CREDSYNCER_VERSION=$(shell source hack/scripts/version; echo $${CREDSYNCER_VERSION})
CREDSYNCER_IMAGE=$(shell source hack/scripts/version; echo $${CREDSYNCER_IMAGE})
bin/credsyncer: pkg/apis/refunc/v1/*.go pkg/credsyncer/*.go cmd/credsyncer/*.go
credsyncer-image: IMAGE=$(CREDSYNCER_IMAGE)

versions: images
	@echo 'controller: ' >values.images.yaml; \
	echo '  image: bundle:${BUNDLE_VERSION}' >>values.images.yaml; \
	echo 'credsyncer: ' >>values.images.yaml; \
	echo '  image: credsyncer:${CREDSYNCER_VERSION}' >>values.images.yaml; \
	echo 'triggers: ' >>values.images.yaml; \
	echo '  eventTrigger: ' >>values.images.yaml; \
	echo '    image: bundle:${BUNDLE_VERSION}' >>values.images.yaml; \
	echo '  timeTrigger: ' >>values.images.yaml; \
	echo '    image: bundle:${BUNDLE_VERSION}' >>values.images.yaml;
