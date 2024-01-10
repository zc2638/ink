tag ?= latest
packages = `go list ./... | grep -v github.com/zc2638/ink/test`

build-%:
	@CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix cgo -o _output/$* ./cmd/$*

docker-build-%:
	@docker build -t zc2638/$* -f docker/$*.dockerfile .

docker-tag-%:
	@docker tag zc2638/$* zc2638/$*:$(tag)
	@echo "TAG $(*) $(tag) success"

docker-push-%:
	@docker push zc2638/$*:$(tag)

build: build-inkd build-inker build-inkctl
docker-build: docker-build-inkd docker-build-inker
docker-tag:   docker-tag-inkd   docker-tag-inker
docker-push:  docker-push-inkd  docker-push-inker

clean:
	@rm -rf _output
	@echo "clean complete"

tests:
	@go test $(packages)

e2e:
	@ginkgo -v test/e2e/suite

up:
	@docker compose -f docker/compose.yml up

down:
	@docker compose -f docker/compose.yml down
