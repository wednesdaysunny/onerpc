package eco

import (
	"fmt"
	"strings"
)

const (
	DirectScheme    = "direct"
	DiscovScheme    = "discov"
	EndpointSepChar = ','
	subsetSize      = 32
)

var (
	EndpointSep = fmt.Sprintf("%c", EndpointSepChar)
)

func BuildDirectTarget(endpoints []string) string {
	return fmt.Sprintf("%s:///%s", DirectScheme,
		strings.Join(endpoints, EndpointSep))
}
