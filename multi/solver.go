package multi

import (
	"strings"

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
	zone, err := client.GetHosted(ch.ResolvedZone)
	if err != nil {
		klog.Errorf("Get hosted zone %v error: %v", ch.ResolvedZone, err)
		return err
	}
	rr := ExtractRR(ch.ResolvedFQDN, zone)
	if _, val, err := client.GetRecord(zone, rr, "TXT"); val == ch.Key {
		klog.Infof("get txt record %v.%v already exists: %v", rr, zone, val)
		return nil // already exists
	} else if err != nil && strings.Contains(err.Error(), "not exist") {
		// pass
	} else if err != nil {
		klog.Errorf("Get txt record %v.%v error: %v", rr, zone, err)
		return err
	} else if val != "" {
		klog.Errorf("Get txt record %v.%v already exists: %v - %v", rr, zone, val, ch.Key)
		return errors.Errorf("Get txt record already exists: %v - %v", val, ch.Key)
	}
	// add record
	if err := client.AddRecord(zone, rr, ch.Key, "TXT"); err != nil {
		klog.Errorf("Add txt record %q error: %v", ch.ResolvedFQDN, err)
		return err
	}
	klog.Infof("Presented txt record %v: %v", ch.ResolvedFQDN, ch.Key)
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
	zone, err := client.GetHosted(ch.ResolvedZone)
	if err != nil {
		klog.Errorf("Get hosted zone %v error: %v", ch.ResolvedZone, err)
		return err
	}
	rr := ExtractRR(ch.ResolvedFQDN, zone)
	id, val, err := client.GetRecord(zone, rr, "TXT")
	if err != nil {
		klog.Errorf("Get txt record %v.%v error: %v", rr, zone, err)
		return err
	}
	if val != ch.Key {
		klog.Errorf("Records value does not match: %v", ch.ResolvedFQDN)
		return errors.New("record value does not match")
	}
	// del record
	if err := client.DelRecord(zone, id); err != nil {
		klog.Errorf("Delete txt record %v error: %v", ch.ResolvedFQDN, err)
		return err
	}
	klog.Infof("Cleaned up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	return nil
}
