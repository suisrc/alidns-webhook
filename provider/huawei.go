package provider

import (
	"encoding/json"
	"fmt"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	huawei "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
	"github.com/suisrc/webhook-dns/multi"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var _ multi.Client = (*HuaweiClient)(nil)

type HuaweiClient struct {
	dnsc *huawei.DnsClient
}

type HuaweiConfig struct {
	Region    string                   `json:"region,omitempty"`
	AccessRef cmmeta.SecretKeySelector `json:"accessRef"`
	SecretRef cmmeta.SecretKeySelector `json:"secretRef"`
}

func NewHuawei(cl *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (multi.Client, error) {
	cfg := HuaweiConfig{}
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
	auth, err := basic.NewCredentialsBuilder().WithAk(string(access)).WithSk(string(secret)).SafeBuild()
	if err != nil {
		klog.Errorf("error creating huawei auth: %v", err)
		return nil, err
	}
	regx, err := region.SafeValueOf(cfg.Region)
	if err != nil {
		klog.Errorf("error creating huawei region: %v", err)
		return nil, err
	}
	hcc, err := huawei.DnsClientBuilder().WithRegion(regx).WithCredential(auth).SafeBuild()
	if err != nil {
		klog.Errorf("error creating huawei client: %v", err)
		return nil, err
	}
	client := huawei.NewDnsClient(hcc)
	return &HuaweiClient{dnsc: client}, nil
}

func (aa *HuaweiClient) GetHosted(zone string) (string, error) {
	req := &model.ListPublicZonesRequest{}
	limit := int32(1)
	req.Limit = &limit
	req.Name = &zone
	zones, err := aa.dnsc.ListPublicZones(req)
	if err != nil {
		return "", fmt.Errorf("unable to get zone: %w", err)
	}
	for _, z := range *zones.Zones {
		if *z.Name == zone {
			return *z.Id, nil
		}
	}
	return "", fmt.Errorf("zone %s does not exist", zone)
	// return util.UnFqdn(zone), nil
}

func (aa *HuaweiClient) AddRecord(zone, rr, val, typ string) error {
	ttl := int32(300)
	req := model.CreateRecordSetRequest{}
	req.ZoneId = zone
	req.Body = &model.CreateRecordSetRequestBody{
		Type:    typ,
		Ttl:     &ttl,
		Name:    rr,
		Records: []string{val},
	}
	_, err := aa.dnsc.CreateRecordSet(&req)
	return err
}

func (aa *HuaweiClient) GetRecord(zone, rr, typ string) (any, string, error) {
	req := model.ListRecordSetsByZoneRequest{}
	req.ZoneId = zone
	req.Type = &typ
	req.Name = &rr

	resp, err := aa.dnsc.ListRecordSetsByZone(&req)
	if err != nil {
		return "", "", err
	}

	var record *model.ListRecordSets = nil
	for _, r := range *resp.Recordsets {
		if *r.Name == rr {
			record = &r
			break
		}
	}

	if record == nil {
		return "", "", ErrNoRecord
	}
	return *record.Id, (*record.Records)[0], nil
}

func (aa *HuaweiClient) DelRecord(zone string, id any) error {
	req := model.DeleteRecordSetRequest{}
	req.ZoneId = zone
	req.RecordsetId = id.(string)
	_, err := aa.dnsc.DeleteRecordSet(&req)
	return err
}
