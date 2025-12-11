package multi

import (
	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"k8s.io/client-go/rest"
)

func NewSolver() webhook.Solver {
	return &MultiSolver{}
}

var _ webhook.Solver = (*MultiSolver)(nil)

type MultiSolver struct {
}

func (c *MultiSolver) Name() string {
	return "suisrc"
}

func (c *MultiSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	return nil
}

func (c *MultiSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	return nil
}

func (c *MultiSolver) Initialize(kc *rest.Config, stopCh <-chan struct{}) error {
	return nil
}
