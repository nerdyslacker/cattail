package ts

import (
	"log/slog"
	"os/user"

	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnstate"
)

type Status struct {
	Status *ipnstate.Status
	Prefs  *ipn.Prefs
}

func (s Status) Online() bool {
	return (s.Status != nil) && (s.Status.BackendState == ipn.Running.String())
}

func (s Status) NeedsLogin() bool {
	return (s.Status != nil) && (s.Status.BackendState == ipn.NeedsLogin.String())
}

func (s Status) OperatorIsCurrent() bool {
	current, err := user.Current()
	if err != nil {
		slog.Error("get current user", "err", err)
		return false
	}

	return s.Prefs.OperatorUser == current.Username
}
