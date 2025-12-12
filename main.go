package main

import (
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/suisrc/webhook-dns/multi"

	// This will register the provider with the webhook serving library.
	_ "github.com/suisrc/webhook-dns/provider"
)

func main() {
	if multi.GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(multi.GroupName, multi.NewSolver())
}
