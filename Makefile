GOOS ?= linux
GOARCH ?= amd64
tag ?= latest

build-%:
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix cgo -o output/$* ./cmd/$*

docker-build-%: build-%
	@docker build --platform $(GOOS)/$(GOARCH) -t zc2638/$* -f build/Dockerfile .

docker-tag-%:
	@docker tag zc2638/$* zc2638/$*:$(tag)
	@echo "TAG $(*) $(tag) success"

docker-push-%:
	@docker push zc2638/$*:$(tag)

build: build-inkd build-inker build-inkctl
docker-build: docker-build-inkd docker-build-inker docker-build-inkctl
docker-tag:   docker-tag-inkd   docker-tag-inker   docker-tag-inkctl
docker-push:  docker-push-inkd  docker-push-inker  docker-push-inkctl

clean:
	@rm -rf output
	@echo "clean complete"

test:
	@go test $(go list ./... | grep -v github.com/zc2638/ink/test)

e2e:
	@ginkgo -v test/e2e/suite
