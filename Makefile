NAME=kubectl-doctor
PACKAGE_NAME=github.com/emirozer/$(NAME)
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)
TAG=$(shell git describe --abbrev=0 --tags)


all: build

$(GOPATH)/bin/golint$(suffix):
	go get github.com/golang/lint/golint

$(GOPATH)/bin/goveralls$(suffix):
	go get github.com/mattn/goveralls

vendor:
	go mod vendor

bin:
	mkdir bin

dep:
	glide up -v

release:
	goreleaser --rm-dist

snapshot:
	goreleaser --snapshot --skip-publish --rm-dist

build: bin
	go build -o kubectl-doctor cmd/plugin/main.go 
	cp ./kubectl-doctor /usr/local/bin/plugins
	cp ./kubectl-doctor ./bin

lint: $(GOPATH)/bin/golint$(suffix)
	golint

vet:
	go vet

test: vet
	go test -race -v -cover ./...

watch:
	ls */*.go | entr make test

cover:
	go test -v ./$(NAME) -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

clean:
	rm -fr dist bin
	rm /usr/local/bin/plugins/kubectl-doctor

fmt:
	gofmt -w $(GOFMT_FILES)

dist/$(NAME)-checksum-%:
	cd dist && sha256sum $@.zip

checksums: dist/$(NAME)-checksum-darwin-amd64 dist/$(NAME)-checksum-windows-386 dist/$(NAME)-checksum-windows-amd64 dist/$(NAME)-checksum-linux-amd64

chocolatey/$(NAME)/$(NAME).$(TAG).nupkg: chocolatey/$(NAME)/$(NAME).nuspec
	cd chocolatey/$(NAME) && choco pack

choco:
	cd chocolatey/$(NAME) && choco push $(NAME).$(TAG).nupkg -s https://chocolatey.org/

.PHONY: release snapshot fmt clean cover acceptance lint test vet watch build check choco checksums
