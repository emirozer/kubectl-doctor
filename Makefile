NAME=kubectl-doctor
PACKAGE_NAME=github.com/emirozer/$(NAME)
TAG=$(shell git describe --abbrev=0 --tags)


all: build

$(GOPATH)/bin/golint$(suffix):
	go get github.com/golang/lint/golint

$(GOPATH)/bin/goveralls$(suffix):
	go get github.com/mattn/goveralls

bin:
	mkdir bin

dep:
	glide up -v

build: bin
	go build -o kubectl-doctor cmd/kubectl-doctor.go
	cp ./kubectl-doctor /usr/local/bin/plugins
	mv ./kubectl-doctor ./bin

lint: $(GOPATH)/bin/golint$(suffix)
	golint

vet:
	go vet

test: vet
	go test -race -v -cover ./...

clean:
	rm -fr dist bin
	rm /usr/local/bin/plugins/kubectl-doctor

fmt:
	gofmt -w $(GOFMT_FILES)

release:
	mkdir dist
	env GOOS=linux GOARCH=amd64 go build -o kubectl-doctor cmd/kubectl-doctor.go ; mv kubectl-doctor dist/kubectl-doctor_linux_amd64
	env GOOS=linux GOARCH=arm go build -o kubectl-doctor cmd/kubectl-doctor.go ; mv kubectl-doctor dist/kubectl-doctor_linux_arm
	env GOOS=darwin GOARCH=amd64 go build -o kubectl-doctor cmd/kubectl-doctor.go ; mv kubectl-doctor dist/kubectl-doctor_darwin_amd64
	env GOOS=windows GOARCH=amd64 go build -o kubectl-doctor cmd/kubectl-doctor.go ; mv kubectl-doctor dist/kubectl-doctor_windows_amd64.exe
	env GOOS=freebsd GOARCH=amd64 go build -o kubectl-doctor cmd/kubectl-doctor.go ; mv kubectl-doctor dist/kubectl-doctor_freebsd_amd64

dist/$(NAME)-checksum-%:
	cd dist && sha256sum $@.zip

checksums: dist/$(NAME)-checksum-darwin-amd64 dist/$(NAME)-checksum-windows-386 dist/$(NAME)-checksum-windows-amd64 dist/$(NAME)-checksum-linux-amd64


.PHONY: fmt clean lint build
