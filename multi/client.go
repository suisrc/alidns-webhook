package multi

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	cmmeta "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type DnsBuilderFunc func(cl *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest) (DnsClient, error)

var DnsBuilders = map[string]DnsBuilderFunc{}

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

type Config struct {
	Provider  string                   `json:"provider,omitempty"`
	AccessRef cmmeta.SecretKeySelector `json:"accessRef,omitempty"`
	SecretRef cmmeta.SecretKeySelector `json:"secretRef,omitempty"`
}

type IConfig interface {
	GetConfig() *Config
}

func (aa *MultiSolver) newClient(ch *v1alpha1.ChallengeRequest) (DnsClient, error) {
	cfg := &Config{}
	// handle the 'base case' where no configuration has been provided
	if err := json.Unmarshal(ch.Config.Raw, cfg); err != nil {
		return nil, errors.Wrap(err, "error decoding solver config")
	} else if cfg.Provider == "" {
		// return nil, errors.New("solver no provider")
		cfg.Provider = "alidns" // default
	}
	if cb, ok := DnsBuilders[cfg.Provider]; ok {
		return cb(aa.client, ch)
	}
	return nil, errors.New("solver no provider: " + cfg.Provider)
}

func LoadConfig(cl *kubernetes.Clientset, ch *v1alpha1.ChallengeRequest, cf IConfig) (access, secret []byte, err error) {
	if err = json.Unmarshal(ch.Config.Raw, cf); err != nil {
		err = errors.Errorf("error decoding solver config: %v", err)
		klog.Error(err.Error()) // log
		return
	}
	klog.Infof("decoded config: %v", cf)
	if cc := cf.GetConfig(); cc == nil {
		// pass
	} else if access, err = GetSecretData(cl, cc.AccessRef, ch.ResourceNamespace); err != nil {
		klog.Errorf("error getting secret %s/%s: %v", ch.ResourceNamespace, cc.AccessRef.Name, err)
	} else if secret, err = GetSecretData(cl, cc.SecretRef, ch.ResourceNamespace); err != nil {
		klog.Errorf("error getting secret %s/%s: %v", ch.ResourceNamespace, cc.SecretRef.Name, err)
	}
	return
}

func ExtractRR(fqdn, domain string) string {
	name := util.UnFqdn(fqdn)
	if idx := strings.Index(name, "."+domain); idx != -1 {
		return name[:idx]
	}
	return name
}

func GetSecretData(client *kubernetes.Clientset, selector cmmeta.SecretKeySelector, ns string) ([]byte, error) {
	secret, err := client.CoreV1().Secrets(ns).Get(context.TODO(), selector.Name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load secret %q", ns+"/"+selector.Name)
	}
	if data, ok := secret.Data[selector.Key]; ok {
		return data, nil
	}
	return nil, errors.Errorf("no key %q in secret %q", selector.Key, ns+"/"+selector.Name)
}
