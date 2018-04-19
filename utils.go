package main

import (
	"fmt"
	"net"

	"github.com/juliengk/go-utils/validation"
)

func HostLookup(name string) (string, error) {
	if err := validation.IsValidIP(name); err == nil {
		return name, nil
	}

	if err := validation.IsValidFQDN(name); err != nil {
		return "", err
	}

	addrs, err := net.LookupHost(name)
	if err != nil {
		return "", err
	}

	if len(addrs) > 0 {
		return addrs[0], nil
	}

	return "", fmt.Errorf("%s cannot be resolved", name)
}
