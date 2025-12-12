package provider

import (
	"encoding/json"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/suisrc/webhook-dns/multi"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var _ multi.Client = (*AlidnsClient)(nil)

type AlidnsClient struct {
	dnsc *alidns.Client
}

type AlidnsConfig struct {
	Region    string                   `json:"region,omitempty"`
	AccessRef cmmeta.SecretKeySelector `json:"accessRef"`
	SecretRef cmmeta.SecretKeySelector `json:"secretRef"`
}

func NewAlidns(cl *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (multi.Client, error) {
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

// func init() {
// 	multi.ClientBuilders["alidns"] = NewAlidns
// 	klog.Info("Registered alidns provider, suisrc/webhook-dns")
// }
