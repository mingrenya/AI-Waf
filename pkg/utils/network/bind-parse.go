package network

import (
	"net/url"
)

func NetworkAddressFromBind(bind string) (network string, address string) {
	bindUrl, err := url.Parse(bind)
	if err == nil {
		return bindUrl.Scheme, bindUrl.Path
	}

	return "tcp", bind
}
