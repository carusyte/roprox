package fetcher

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/carusyte/roprox/types"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/pkg/errors"
)

//SpysOne fetches proxy server from http://spys.one
type SpysOne struct {
	URLs []string
}

//UID returns the unique identifier for this spec.
func (f SpysOne) UID() string {
	return "SpysOne"
}

//Urls return the server urls that provide the free proxy server lists.
func (f SpysOne) Urls() []string {
	if len(f.URLs) > 0 {
		return f.URLs
	}
	return []string{
		`http://spys.one/en/anonymous-proxy-list/`,
		`http://spys.one/en/socks-proxy-list/`,
	}
}

//UseMasterProxy returns whether the fetcher needs a master proxy server
//to access the free proxy list provider.
func (f SpysOne) UseMasterProxy() bool {
	return true
}

//RefreshInterval determines how often the list should be refreshed, in minutes.
func (f SpysOne) RefreshInterval() int {
	return 30
}

//Fetch the proxy info.
func (f SpysOne) Fetch(ctx context.Context, urlIdx int, url string) (ps []*types.ProxyServer, e error) {
	ipPort := make([]string, 0, 4)
	ts := make([]string, 0, 4)
	anon := make([]string, 0, 4)
	locations := make([]string, 0, 4)
	var max, xppLen int
	var str string

	if e = chromedp.Run(ctx,
		chromedp.WaitReady(`#xpp`),
		chromedp.JavascriptAttribute(`#xpp`, `length`, &xppLen),
		chromedp.TextContent(`#xpp option:last-child`, &str),
		chromedp.SetAttributeValue(`#xpp`, "multiple", ""),
		chromedp.SetAttributeValue(`#xpp`, "size", strconv.Itoa(xppLen)),
		chromedp.Click(`#xpp option:last-child`),
	); e != nil {
		e = errors.Wrapf(e, "failed to manipulate #xpp")
		return ps, e
	}
	log.Debugf("#xpp len: %d, max record string: %s", xppLen, str)
	if max, e = strconv.Atoi(str); e != nil {
		return ps, errors.Wrapf(e, "unable to convert max record string: %s", str)
	}

	if e = chromedp.Run(ctx,
		chromedp.WaitReady(`body > table:nth-child(3) > tbody > tr:nth-child(5) > td > table`),
		chromedp.WaitReady(`body > table:nth-child(3) > tbody > tr:nth-child(5) > td `+
			`> table > tbody > tr:nth-child(30)`),
	); e != nil {
		return ps, errors.Wrapf(e, "failed to wait page refresh")
	}

	if e = f.pageEnd(ctx); e != nil {
		return
	}

	typeSel := `body > table:nth-child(3) > tbody > tr:nth-child(5) ` +
		`> td > table > tbody > tr > td:nth-child(2) > a`
	if strings.Contains(url, "socks-proxy-list") {
		typeSel = `body > table:nth-child(3) > tbody > tr:nth-child(5) ` +
			`> td > table > tbody > tr:not(:nth-child(2)) > td:nth-child(2)`
	}

	if e = chromedp.Run(ctx,
		chromedp.WaitReady(fmt.Sprintf(`body > table:nth-child(3) > tbody > tr:nth-child(5) > td `+
			`> table > tbody > tr:nth-child(%d)`, max)),
		//get ip and port
		chromedp.Evaluate(jsGetText(`body > table:nth-child(3) > tbody > tr:nth-child(5) `+
			`> td > table > tbody > tr > td:nth-child(1) > font.spy14`), &ipPort),
		//get types
		chromedp.Evaluate(jsGetText(typeSel), &ts),
		//get anonymity
		chromedp.Evaluate(jsGetText(`body > table:nth-child(3) > tbody > tr:nth-child(5) `+
			`> td > table > tbody > tr > td:nth-child(3) > a > font`), &anon),
		//get location
		chromedp.Evaluate(jsGetText(`body > table:nth-child(3) > tbody > tr:nth-child(5) `+
			`> td > table > tbody > tr > td:nth-child(4)`), &locations),
	); e != nil {
		return ps, errors.Wrapf(e, "failed to extract proxy info")
	}

	return f.parse(ipPort, ts, anon, locations), nil
}

func (f SpysOne) pageEnd(ctx context.Context) (e error) {
	var bottom bool
	for i := 1; true; i++ {
		if e = chromedp.Run(ctx,
			chromedp.KeyEvent(kb.End),
		); e != nil {
			return errors.Wrapf(e, "failed to send kb.End key #%d", i)
		}

		log.Debugf("End key sent #%d", i)

		if e = chromedp.Run(ctx,
			chromedp.Evaluate(jsPageBottom(), &bottom),
		); e != nil {
			return errors.Wrapf(e, "failed to check page bottom #%d", i)
		}

		if bottom {
			//found footer
			break
		}

		time.Sleep(time.Millisecond * 500)
	}
	return
}

//parses the selected values into proxy server instances
func (f SpysOne) parse(ipPort, ts, anon, locations []string) (ps []*types.ProxyServer) {
	for i, d := range ipPort {
		if len(anon) <= i {
			break
		}
		if len(locations) <= i {
			break
		}
		if len(ts) <= i {
			break
		}

		a := strings.TrimSpace(anon[i])
		if strings.EqualFold(a, "NOA") {
			//non anonymous proxy
			continue
		}

		ss := strings.Split(strings.TrimSpace(d), ":")
		if len(ss) != 2 {
			log.Warnf("%s possible invalid ip & port string, skipping: %+v", f.UID(), d)
			continue
		}
		host, port := strings.TrimSpace(ss[0]), strings.TrimSpace(ss[1])

		t := strings.ToLower(strings.TrimSpace(ts[i]))
		if strings.Contains(t, "http") {
			t = "http"
		} else if strings.Contains(t, "socks5") {
			t = "socks5"
		} else {
			log.Debugf("%s unsupported proxy type: %+v", f.UID(), t)
			continue
		}

		loc := strings.TrimSpace(strings.ReplaceAll(locations[i], "!", ""))

		ps = append(ps, &types.ProxyServer{
			Source: f.UID(),
			Host:   host,
			Port:   port,
			Type:   t,
			Loc:    loc,
		})
	}
	return
}
