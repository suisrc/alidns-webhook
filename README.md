<p align="center">
  <img src="https://raw.githubusercontent.com/cert-manager/cert-manager/d53c0b9270f8cd90d908460d69502694e1838f5f/logo/logo-small.png" height="256" width="256" alt="cert-manager project logo" />
</p>

# ACME webhook

## alidns



## dnspod



## huawei

## 测试环境

```sh
# https://go.kubebuilder.io/test-tools/
curl -sL  https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-v1.19.2-linux-amd64.tar.gz -o _test/kubebuilder.tar.gz
tar -xzvf testbin/kubebuilder.tar.gz -C _test

# install kubebuilder tools bin to default cert-manager tools path
mkdir -p ~/go/pkg/mod/github.com/cert-manager/cert-manager@v1.19.2/_bin/tools
cp -r _test/kubebuilder/bin/* ~/go/pkg/mod/github.com/cert-manager/cert-manager@v1.19.2/_bin/tools
```


## 感谢

[webhook-example](https://github.com/cert-manager/webhook-example)
[cert-manager-webhook-dnspod](https://github.com/imroc/cert-manager-webhook-dnspod)
