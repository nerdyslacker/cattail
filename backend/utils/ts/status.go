package ts

import (
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
)

type Status struct {
	Status *ipnstate.Status
}

func (s Status) Online() bool {
	return (s.Status != nil) && (s.Status.BackendState == ipn.Running.String())
}

func (s Status) NeedsAuth() bool {
	return (s.Status != nil) && (s.Status.BackendState == ipn.NeedsLogin.String())
}
