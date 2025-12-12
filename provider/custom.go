package provider

import (
	"fmt"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/suisrc/webhook-dns/multi"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var _ multi.DnsClient = (*CustomClient)(nil)

type CustomClient struct {
	conf CustomConfig
}

type CustomConfig struct {
	multi.Config
	Region string `json:"region,omitempty"`

	access string `json:"-"`
	secret string `json:"-"`
}

func (cc *CustomConfig) GetConfig() *multi.Config {
	return &cc.Config
}

var (
	_custom_records map[string]string = nil
)

func NewCustom(cl *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (multi.DnsClient, error) {
	if _custom_records == nil {
		_custom_records = make(map[string]string)
	}
	cfg := CustomConfig{}
	access, secret, err := multi.LoadConfig(cl, ch, &cfg)
	if err != nil {
		return nil, err
	}
	cfg.access = string(access)
	cfg.secret = string(secret)
	return &CustomClient{conf: cfg}, nil
}

func (aa *CustomClient) GetHosted(zone string) (string, error) {
	return util.UnFqdn(zone), nil
}

func (aa *CustomClient) AddRecord(zone, rr, val, typ string) error {
	klog.Infof("[%v-%v]AddRecord: %v.%v %v %v", aa.conf.access, aa.conf.secret, rr, zone, val, typ)
	_custom_records[fmt.Sprintf("%v.%v", rr, zone)] = val
	return nil
}

func (aa *CustomClient) GetRecord(zone, rr, typ string) (any, string, error) {
	klog.Infof("[%v-%v]GetRecord: %v.%v %v", aa.conf.access, aa.conf.secret, rr, zone, typ)
	id := fmt.Sprintf("%v.%v", rr, zone)
	val, ok := _custom_records[id]
	if !ok {
		return "", "", ErrNoRecord
	}
	return id, val, nil
}

func (aa *CustomClient) DelRecord(zone string, id any) error {
	klog.Infof("[%v-%v]DelRecord: %v", aa.conf.access, aa.conf.secret, id)
	delete(_custom_records, id.(string))
	return nil
}
