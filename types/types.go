package types

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/carusyte/roprox/logging"
)

var log = logging.Logger

type ContentType string

const (
	//OK indicates the proxy server is available per the last check result
	OK = "OK"
	//FAIL indicates the proxy server is unavailable per the last check result
	FAIL = "FAIL"
	//UNK indicators the proxy server status is unknown.
	UNK = "UNK"
	//DateTimeFormat the default date time format string in go
	DateTimeFormat = "2006-01-02 15:04:05"
)

//ProxyServer is a model mapping for database table proxy_list
type ProxyServer struct {
	Source      string
	Host        string
	Port        string
	Type        string
	Loc         string
	Status      string
	Suc         int
	Fail        int
	Score       float64
	LastCheck   string `db:"last_check"`
	LastScanned string `db:"last_scanned"`
}

func (p *ProxyServer) String() string {
	j, e := json.Marshal(p)
	if e != nil {
		log.Error(e)
	}
	return fmt.Sprintf("%v", string(j))
}

//NewProxyServer creates an instance of ProxyServer.
func NewProxyServer(source, host, port, stype string) *ProxyServer {
	return &ProxyServer{
		Source:      source,
		Host:        host,
		Port:        port,
		Type:        stype,
		Status:      UNK,
		Fail:        0,
		LastCheck:   time.Now().Format(DateTimeFormat),
		LastScanned: time.Now().Format(DateTimeFormat),
	}
}

//FetcherSpec defines detail specifications on fetching open proxy servers from the web.
type FetcherSpec interface {
	//UID returns the unique identifier for this spec.
	UID() string
	//Urls return the server urls that provide the free proxy server lists.
	Urls() []string
	//UseMasterProxy returns whether the fetcher needs a master proxy server
	//to access the free proxy list provider.
	UseMasterProxy() bool
	//RefreshInterval determines how often the list should be refreshed, in minutes.
	RefreshInterval() int
}

//JSONFetcher parses target url as JSON payload
type JSONFetcher interface {
	//ParseJSON parses JSON payload and extracts proxy information
	ParseJSON(payload []byte) (ps []*ProxyServer)
}

//StaticHTMLFetcher fetches target url by parsing static HTML content
type StaticHTMLFetcher interface {
	//IsGBK returns wheter the web page is GBK encoded.
	IsGBK() bool
	//ListSelector returns the jQuery selectors for searching the proxy server list/table.
	ListSelector() []string
	//ScanItem process each item found in the table determined by ListSelector().
	ScanItem(itemIdx, urlIdx int, s *goquery.Selection) (ps *ProxyServer)
}

//DynamicHTMLFetcher fetches target url by using web driver
type DynamicHTMLFetcher interface {
	//Fetch the proxy info
	Fetch(ctx context.Context, urlIdx int, url string ) (ps []*ProxyServer, e error)
}

//UserAgent represents user_agent table structure.
type UserAgent struct {
	ID                   string
	UserAgent            string `db:"user_agent"`
	TimesSeen            string `db:"times_seen"`
	SimpleSoftwareString string `db:"simple_software_string"`
	SoftwareName         string `db:"software_name"`
	SoftwareVersion      string `db:"software_version"`
	SoftwareType         string `db:"software_type"`
	SoftwareSubType      string `db:"software_sub_type"`
	HardWareType         string `db:"hardware_type"`
	FirstSeenAt          string `db:"first_seen_at"`
	LastSeenAt           string `db:"last_seen_at"`
	UpdatedAt            string `db:"updated_at"`
}

func (ua *UserAgent) String() string {
	j, e := json.Marshal(ua)
	if e != nil {
		log.Error(e)
	}
	return fmt.Sprintf("%v", string(j))
}
