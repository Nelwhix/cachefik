package docker

import "strings"

type Service struct {
	Rule     string
	Upstream string
}

func (s Service) PathPrefix() string {
	return strings.TrimSuffix(strings.TrimPrefix(s.Rule, "PathPrefix(`"), "`)")
}
