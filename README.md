<p align="center">
  <img src="https://raw.githubusercontent.com/cert-manager/cert-manager/d53c0b9270f8cd90d908460d69502694e1838f5f/logo/logo-small.png" height="256" width="256" alt="cert-manager project logo" />
</p>

# ACME webhook

多 webhook 实现， 目前支持：  
- alidns: 阿里(默认，provider: "")
- dnspod: 腾讯
- huawei: 华为(暂未测试)

## alidns

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: as-alidns
  namespace: cert-manager
stringData:
  access: LT...
  secret: HA...

---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer 
metadata:
  name: letsencrypt-alidns
  namespace: cert-manager
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: mail@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    # disableAccountKeyGeneration: true
    # Enable the DNS-01 challenge provider
    solvers:
    - dns01:
        webhook:
          solverName: suisrc
          groupName: acme.suisrc.com
          provider: alidns
          config:
            region: ""
            accessRef:
              name: as-alidns
              key: access
            secretRef:
              name: as-alidns
              key: secret

```

## dnspod

```yaml

apiVersion: v1
kind: Secret
metadata:
  name: as-dnspod
  namespace: cert-manager
stringData:
  access: AK...
  secret: HA...

---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: dnspod
  namespace: cert-manager
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: mail@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    # disableAccountKeyGeneration: true
    # Enable the DNS-01 challenge provider
    solvers:
    - dns01:
        webhook:
          solverName: suisrc
          groupName: acme.suisrc.com
          provider: alidns
          config:
            ttl: 600
            recordLine: ""
            accessRef:
              name: as-dnspod
              key: access
            secretRef:
              name: as-dnspod
              key: secret

```

## huawei

没有账号，暂时没有测试

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: as-huawei
  namespace: cert-manager
stringData:
  access: LT...
  secret: HA...

---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer 
metadata:
  name: letsencrypt-huawei
  namespace: cert-manager
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: mail@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    # disableAccountKeyGeneration: true
    # Enable the DNS-01 challenge provider
    solvers:
    - dns01:
        webhook:
          solverName: suisrc
          groupName: acme.suisrc.com
          provider: huawei
          config:
            region: "cn-north-4"
            accessRef:
              name: as-huawei
              key: access
            secretRef:
              name: as-huawei
              key: secret

```

## 扩展DNS服务商

实现DNS服务商，需要实现 DNSClient 接口
```go
type DnsClient interface {
	// get hosted
	GetHosted(zone string) (string, error)
	// add record
	AddRecord(zone, rr, val, typ string) error
	// get record
	GetRecord(zone, rr, typ string) (any, string, error)
	// del record
	DelRecord(zone string, id any) error
}
```

例子 alidns
```go

var _ multi.DnsClient = (*AlidnsClient)(nil)

type AlidnsClient struct {
	dnsc *alidns.Client
}

type AlidnsConfig struct {
	Region    string                   `json:"region,omitempty"`
	AccessRef cmmeta.SecretKeySelector `json:"accessRef"`
	SecretRef cmmeta.SecretKeySelector `json:"secretRef"`
}

func NewAlidns(cl *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (multi.DnsClient, error) {
	cfg := AlidnsConfig{}
	if err := json.Unmarshal(ch.Config.Raw, &cfg); err != nil {
		return nil, fmt.Errorf("error decoding solver config: %v", err)
	}
	klog.Infof("Decoded config: %v", cfg)
	access, err := multi.GetSecretData(cl, cfg.AccessRef, ch.ResourceNamespace)
	if err != nil {
		klog.Errorf("error getting secret %s/%s: %v", ch.ResourceNamespace, cfg.AccessRef.Name, err)
		return nil, err
	}
	secret, err := multi.GetSecretData(cl, cfg.SecretRef, ch.ResourceNamespace)
	if err != nil {
		klog.Errorf("error getting secret %s/%s: %v", ch.ResourceNamespace, cfg.SecretRef.Name, err)
		return nil, err
	}

	cred := credentials.NewAccessKeyCredential(string(access), string(secret))
	client, err := alidns.NewClientWithOptions(cfg.Region, sdk.NewConfig(), cred)
	if err != nil {
		klog.Errorf("error creating alidns client: %v", err)
		return nil, err
	}
	return &AlidnsClient{dnsc: client}, nil
}

func (aa *AlidnsClient) GetHosted(zone string) (string, error) {
	req := alidns.CreateDescribeDomainsRequest()
	req.KeyWord = util.UnFqdn(zone)
	req.SearchMode = "EXACT"

	resp, err := aa.dnsc.DescribeDomains(req)
	if err != nil {
		return "", err
	}

	zones := resp.Domains.Domain
	if len(zones) == 0 {
		return "", fmt.Errorf("zone %s does not exist", zone)
	}

	return zones[0].DomainName, nil
}

func (aa *AlidnsClient) AddRecord(zone, rr, val, typ string) error {
	req := alidns.CreateAddDomainRecordRequest()
	req.DomainName = zone
	req.Type = typ
	req.TTL = "600"
	req.RR = rr
	req.Value = val
	_, err := aa.dnsc.AddDomainRecord(req)
	return err
}

func (aa *AlidnsClient) GetRecord(zone, rr, typ string) (any, string, error) {
	req := alidns.CreateDescribeDomainRecordsRequest()
	req.Type = typ
	req.DomainName = zone
	req.RRKeyWord = rr

	resp, err := aa.dnsc.DescribeDomainRecords(req)
	if err != nil {
		return "", "", err
	}

	var record *alidns.Record = nil
	for _, r := range resp.DomainRecords.Record {
		if r.RR == rr {
			record = &r
			break
		}
	}

	if record == nil {
		return "", "", ErrNoRecord
	}
	return record.RecordId, record.Value, nil
}

func (aa *AlidnsClient) DelRecord(zone string, id any) error {
	req := alidns.CreateDeleteDomainRecordRequest()
	// req.Domain = zone
	req.RecordId = id.(string)
	_, err := aa.dnsc.DeleteDomainRecord(req)
	return err
}

```


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
