package provider

import (
	"strings"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/suisrc/webhook-dns/multi"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var _ multi.DnsClient = (*DnspodClient)(nil)

type DnspodClient struct {
	dnsc *dnspod.Client
	conf DnspodConfig
}

type DnspodConfig struct {
	multi.Config
	TTL        *uint64 `json:"ttl,omitempty"`
	RecordLine string  `json:"recordLine,omitempty"`
}

func (cc *DnspodConfig) GetConfig() *multi.Config {
	return &cc.Config
}

func NewDnspod(cl *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (multi.DnsClient, error) {
	cfg := DnspodConfig{}
	access, secret, err := multi.LoadConfig(cl, ch, &cfg)
	if err != nil {
		return nil, err
	}
	if cfg.RecordLine == "" {
		cfg.RecordLine = "默认"
	}
	if cfg.TTL == nil {
		cfg.TTL = common.Uint64Ptr(600)
	}

	cred := common.NewCredential(string(access), string(secret))
	client, err := dnspod.NewClient(cred, "", profile.NewClientProfile())
	if err != nil {
		klog.Errorf("error creating dnspod client: %v", err)
		return nil, err
	}
	return &DnspodClient{dnsc: client, conf: cfg}, nil
}

func (aa *DnspodClient) GetHosted(zone string) (string, error) {
	// req := dnspod.NewDescribeDomainRequest()
	// req.Domain = common.StringPtr(util.UnFqdn(zone))
	// resp, err := aa.dnsc.DescribeDomain(req)
	// if err != nil {
	// 	return "", err
	// }
	// info := resp.Response.DomainInfo
	// if info == nil {
	// 	return "", fmt.Errorf("zone %s does not exist", zone)
	// }
	// return *info.Domain, nil

	return util.UnFqdn(zone), nil
}

func (aa *DnspodClient) AddRecord(zone, rr, val, typ string) error {
	req := dnspod.NewCreateRecordRequest()
	req.Domain = common.StringPtr(zone)
	req.TTL = aa.conf.TTL
	req.Value = common.StringPtr(val)
	req.RecordType = common.StringPtr(typ)
	req.RecordLine = common.StringPtr(aa.conf.RecordLine)
	req.SubDomain = common.StringPtr(rr)
	_, err := aa.dnsc.CreateRecord(req)
	return err
}

func (aa *DnspodClient) GetRecord(zone, rr, typ string) (any, string, error) {
	req := dnspod.NewDescribeRecordListRequest()
	req.Limit = common.Uint64Ptr(1)
	req.Domain = common.StringPtr(zone)
	// req.Keyword = common.StringPtr(rr)
	req.Subdomain = common.StringPtr(rr)
	req.RecordType = common.StringPtr(typ)
	resp, err := aa.dnsc.DescribeRecordList(req)
	if err != nil {
		if strings.Contains(err.Error(), "ResourceNotFound.NoDataOfRecord") {
			return "", "", ErrNoRecord // record does not exist
		}
		return "", "", err
	}
	var record *dnspod.RecordListItem = nil
	for _, r := range resp.Response.RecordList {
		if *r.Name == rr {
			record = r
			break
		}
	}
	if record == nil {
		return "", "", ErrNoRecord
	}
	return record.RecordId, *record.Value, nil
}

func (aa *DnspodClient) DelRecord(zone string, id any) error {
	req := dnspod.NewDeleteRecordRequest()
	req.Domain = common.StringPtr(zone)
	req.RecordId = id.(*uint64)
	_, err := aa.dnsc.DeleteRecord(req)
	return err
}

// func init() {
// 	multi.ClientBuilders["dnspod"] = NewDnspod
// 	klog.Info("Registered dnspod provider, suisrc/webhook-dns")
// }
