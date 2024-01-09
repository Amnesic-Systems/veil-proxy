package main

import "github.com/coreos/go-iptables/iptables"

const (
	on  = true
	off = false
)

var (
	iptablesRules = [][]string{
		{"nat", "POSTROUTING", "-s", "10.0.0.0/24", "-j", "MASQUERADE"},
		{"filter", "FORWARD", "-i", TunName, "-s", "10.0.0.0/24", "-j", "ACCEPT"},
		{"filter", "FORWARD", "-o", TunName, "-d", "10.0.0.0/24", "-j", "ACCEPT"},
	}
)

// toggleNAT toggles our iptables NAT rules, which ensure that the enclave can
// talk to the Internet.
func toggleNAT(toggle bool) error {
	t, err := iptables.New()
	if err != nil {
		return err
	}

	f := t.AppendUnique
	if toggle == off {
		f = t.DeleteIfExists
	}

	const table, chain, rulespec = 0, 1, 2
	for _, r := range iptablesRules {
		if err := f(r[table], r[chain], r[rulespec:]...); err != nil {
			return err
		}
	}

	return nil
}
