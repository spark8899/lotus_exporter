package lotus

import (
	"github.com/multiformats/go-multiaddr"
	"net/url"
	"regexp"
	"strings"

	manet "github.com/multiformats/go-multiaddr/net"
)

var (
	infoWithToken = regexp.MustCompile("^[a-zA-Z0-9\\-_]+?\\.[a-zA-Z0-9\\-_]+?\\.([a-zA-Z0-9\\-_]+)?:.+$")
)

type apiConnInfo struct {
	addr  string
	token []byte
}

func dialArgs(addr, version string) (string, error) {
	ma, err := multiaddr.NewMultiaddr(addr)
	if err == nil {
		_, addr, err := manet.DialArgs(ma)
		if err != nil {
			return "", err
		}

		return "ws://" + addr + "/rpc/" + version, nil
	}

	_, err = url.Parse(addr)
	if err != nil {
		return "", err
	}
	return addr + "/rpc/" + version, nil
}

func parseApiInfo(s string) apiConnInfo {
	var tok []byte
	if infoWithToken.Match([]byte(s)) {
		sp := strings.SplitN(s, ":", 2)
		tok = []byte(sp[0])
		s = sp[1]
	}

	addr, err := dialArgs(s, "v0")
	if err != nil {
		return apiConnInfo{
			addr:  s,
			token: tok,
		}
	}

	return apiConnInfo{
		addr:  addr,
		token: tok,
	}
}
