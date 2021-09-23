module github.com/suisrc/alidns-webhook

go 1.12

require (
	github.com/aliyun/alibaba-cloud-sdk-go v0.0.0-20190822073329-cd5cf285f2a3
	github.com/jetstack/cert-manager v1.5.3
	github.com/pkg/errors v0.9.1
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/klog v1.0.0
)

//replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190413052642-108c485f896e
//replace github.com/evanphx/json-patch => github.com/evanphx/json-patch v0.0.0-20190203023257-5858425f7550
