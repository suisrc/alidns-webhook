package testdata_test

import (
	"testing"
	"time"

	dns "github.com/cert-manager/cert-manager/test/acme"
	"github.com/suisrc/webhook-dns/multi"
	_ "github.com/suisrc/webhook-dns/provider"
)

func TestDnspod(t *testing.T) {

	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	fixture := dns.NewFixture(multi.NewSolver(),
		dns.SetResolvedZone("magi2002.com."),
		dns.SetManifestPath("../testdata/dnspod"),
		dns.SetAllowAmbientCredentials(false),
		dns.SetUseAuthoritative(false),
		dns.SetPropagationLimit(5*time.Second), // 等待DNS传播时间
		// dns.SetStrict(false),
	)

	fixture.RunBasic(t)
	fixture.RunExtended(t)
}
