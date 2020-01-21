package fetcher

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/carusyte/roprox/types"
)

//SocksProxy fetches proxy server from https://www.socks-proxy.net
type SocksProxy struct{}

//UID returns the unique identifier for this spec.
func (f SocksProxy) UID() string {
	return "SocksProxy"
}

//Urls return the server urls that provide the free proxy server lists.
func (f SocksProxy) Urls() []string {
	return []string{
		`https://www.socks-proxy.net/`,
	}
}

//IsGBK returns wheter the web page is GBK encoded.
func (f SocksProxy) IsGBK() bool {
	return false
}

//UseMasterProxy returns whether the fetcher needs a master proxy server
//to access the free proxy list provider.
func (f SocksProxy) UseMasterProxy() bool {
	return true
}

//ContentType returns the target url's content type
func (f SocksProxy) ContentType() types.ContentType{
	return types.StaticHTML
}
//ParseJSON parses JSON payload and extracts proxy information
func (f SocksProxy) ParseJSON(payload []byte) (ps []*types.ProxyServer){
	return
}

//ListSelector returns the jQuery selector for searching the proxy server list/table.
func (f SocksProxy) ListSelector() []string {
	return []string{
		"#proxylisttable tbody tr",
	}
}

//RefreshInterval determines how often the list should be refreshed, in minutes.
func (f SocksProxy) RefreshInterval() int {
	return 10
}

//ScanItem process each item found in the table determined by ListSelector().
func (f SocksProxy) ScanItem(i, urlIdx int, s *goquery.Selection) (ps *types.ProxyServer) {
	ptype := strings.TrimSpace(s.Find("td:nth-child(5)").Text())
	if !strings.EqualFold("socks5", ptype) {
		return
	}
	anon := strings.TrimSpace(s.Find("td:nth-child(5)").Text())
	if strings.EqualFold(anon, `transparent`) {
		return
	}
	host := strings.TrimSpace(s.Find("td:nth-child(1)").Text())
	port := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
	ps = types.NewProxyServer(f.UID(), host, port, "socks5")
	return
}
