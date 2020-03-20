package exchange

import (
	"net/http"

	"stash.corp.netflix.com/ps/metatron-go/mttls"
)

func BuildHTTPClient(options *Options) (*http.Client, error) {
	checkRedirect := func(req *http.Request, via []*http.Request) error {
		// Do not follow redirects
		return http.ErrUseLastResponse
	}
	if options.FollowRedirects {
		checkRedirect = nil
	}

	if options.AppName != "" {
		cl, err := mttls.NewHTTPClient(options.AppName)
		if err != nil {
			return nil, err
		}
		cl.CheckRedirect = checkRedirect
		cl.Timeout = options.Timeout

		return cl, nil
	}

	return &http.Client{
		CheckRedirect: checkRedirect,
		Timeout:       options.Timeout,
	}, nil
}
