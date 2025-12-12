.PHONY: start build

NOW = $(shell date -u '+%Y%m%d%I%M%S')

# 初始化mod
init:
	go mod init github.com/suisrc/webhook-dns

# 修正依赖
tidy:
	go mod tidy

build:
	CGO_ENABLED=0 go build -o .tmp/webhook -ldflags '-w -extldflags "-static"' .

# go env -w GOPROXY=https://proxy.golang.com.cn,direct
proxy:
	go env -w GO111MODULE=on
	go env -w GOPROXY=http://mvn.res.local/repository/go,direct
	go env -w GOSUMDB=sum.golang.google.cn

helm:
	helm -n cert-manager template deploy/webhook > deploy/bundle.yml

main:
	GROUP_NAME=suisrc go run main.go

test-custom2:
	TEST_ASSET_ETCD=_test/kubebuilder/bin/etcd \
	TEST_ASSET_KUBE_APISERVER=_test/kubebuilder/bin/kube-apiserver \
	TEST_ASSET_KUBECTL=_test/kubebuilder/bin/kubectl \
	go test -v -run TestCustom testdata/custom_test.go

test-custom:
	go test -v testdata/custom_test.go

test-alidns:
	go test -v testdata/alidns_test.go

test-dnspod:
	go test -v testdata/dnspod_test.go

