package multi

import (
	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var _ webhook.Solver = (*MultiSolver)(nil)

func NewSolver() webhook.Solver {
	return &MultiSolver{}
}

type MultiSolver struct {
	client *kubernetes.Clientset
}

func (aa *MultiSolver) Name() string {
	return "suisrc"
}

func (aa *MultiSolver) Initialize(kc *rest.Config, stopCh <-chan struct{}) error {
	client, err := kubernetes.NewForConfig(kc)
	if err != nil {
		return err
	}
	aa.client = client
	return nil
}

func (aa *MultiSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	if ch == nil {
		klog.Error("solver '" + aa.Name() + "' no configuration")
		return errors.New("solver no configuration")
	}
	klog.Infof("Presenting txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	client, err := aa.newClient(ch)
	if err != nil {
		klog.Errorf("New client from challenge error: %v", err)
		return err
	}
	zone, err := client.GetHostedZone(ch.ResolvedZone)
	if err != nil {
		klog.Errorf("Get hosted zone %v error: %v", ch.ResolvedZone, err)
		return err
	}
	rr := ExtractRR(ch.ResolvedFQDN, ch.ResolvedZone)
	if err := client.AddTxtRecord(zone, rr, ch.Key, "TXT"); err != nil {
		klog.Errorf("Add txt record %q error: %v", ch.ResolvedFQDN, err)
		return err
	}
	klog.Infof("Presented txt record %v", ch.ResolvedFQDN)
	return nil
}

func (aa *MultiSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	if ch == nil {
		klog.Error("solver '" + aa.Name() + "' no configuration")
		return errors.New("solver no configuration")
	}
	klog.Infof("Cleaning up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	client, err := aa.newClient(ch)
	if err != nil {
		klog.Errorf("New client from challenge error: %v", err)
		return err
	}
	zone, err := client.GetHostedZone(ch.ResolvedZone)
	if err != nil {
		klog.Errorf("Get hosted zone %v error: %v", ch.ResolvedZone, err)
		return err
	}
	rr := ExtractRR(ch.ResolvedFQDN, ch.ResolvedZone)
	id, val, err := client.GetTxtRecord(zone, rr, "TXT")
	if err != nil {
		klog.Errorf("Get txt record %v.%v error: %v", rr, zone, err)
		return err
	}
	if val != ch.Key {
		klog.Errorf("Records value does not match: %v", ch.ResolvedFQDN)
		return errors.New("record value does not match")
	}
	if err := client.DelTxtRecord(zone, id); err != nil {
		klog.Errorf("Delete txt record %v error: %v", ch.ResolvedFQDN, err)
		return err
	}
	klog.Infof("Cleaned up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	return nil
}
