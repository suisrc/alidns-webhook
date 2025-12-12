package provider

import (
	"encoding/json"
	"fmt"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/suisrc/webhook-dns/multi"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var _ multi.Client = (*CustomClient)(nil)

type CustomClient struct {
	conf CustomConfig
}

type CustomConfig struct {
	Region    string                   `json:"region,omitempty"`
	AccessRef cmmeta.SecretKeySelector `json:"accessRef"`
	SecretRef cmmeta.SecretKeySelector `json:"secretRef"`

	access string `json:"-"`
	secret string `json:"-"`
}

func NewCustom(cl *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (multi.Client, error) {
	cfg := CustomConfig{}
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
	cfg.access = string(access)
	cfg.secret = string(secret)
	return &CustomClient{conf: cfg}, nil
}

func (aa *CustomClient) GetHosted(zone string) (string, error) {
	return util.UnFqdn(zone), nil
}

func (aa *CustomClient) AddRecord(zone, rr, val, typ string) error {
	klog.Infof("[%v-%v]AddRecord: %v.%v %v %v", aa.conf.access, aa.conf.secret, rr, zone, val, typ)
	return nil
}

func (aa *CustomClient) GetRecord(zone, rr, typ string) (any, string, error) {
	klog.Infof("[%v-%v]GetRecord: %v.%v %v", aa.conf.access, aa.conf.secret, rr, zone, typ)
	id := fmt.Sprintf("%v.%v", rr, zone)
	return id, "123d==", nil
}

func (aa *CustomClient) DelRecord(zone string, id any) error {
	klog.Infof("[%v-%v]DelRecord: %v", aa.conf.access, aa.conf.secret, id)
	return nil
}

// func init() {
// 	multi.ClientBuilders["huawei"] = CustomClient
// 	klog.Info("Registered huawei provider, suisrc/webhook-dns")
// }
