package main

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/carusyte/roprox/conf"
	"github.com/carusyte/roprox/types"
	"github.com/carusyte/roprox/util"
)

func check(wg *sync.WaitGroup) {
	defer wg.Done()

	chjobs := make(chan *types.ProxyServer, 128)
	probe(chjobs)
	collectStaleServers(chjobs)
}

func evictStaleServers() {
	logrus.Debug("evicting stale servers...")
	delete := `delete from proxy_list where status = ? and last_scanned <= ?`
	r, e := db.Exec(delete, types.FAIL, time.Now().Add(
		-time.Duration(conf.Args.EvictionTimeout)*time.Second).Format(util.DateTimeFormat))
	if e != nil {
		logrus.Errorln("failed to evict stale proxy servers", e)
		return
	}
	ra, e := r.RowsAffected()
	if e != nil {
		logrus.Warnf("unable to get rows affected after eviction", e)
		return
	}
	logrus.Debugf("%d stale servers evicted", ra)
}

func collectStaleServers(chjobs chan<- *types.ProxyServer) {
	//kickoff at once and repeatedly
	evictStaleServers()
	queryStaleServers(chjobs)
	probeTk := time.NewTicker(time.Duration(conf.Args.ProbeInterval) * time.Second)
	evictTk := time.NewTicker(time.Duration(conf.Args.EvictionInterval) * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-probeTk.C:
			queryStaleServers(chjobs)
		case <-evictTk.C:
			evictStaleServers()
		case <-quit:
			probeTk.Stop()
			evictTk.Stop()
			return
		}
	}
}

func queryStaleServers(chjobs chan<- *types.ProxyServer) {
	logrus.Debug("collecting stale servers...")
	var list []*types.ProxyServer
	query := `SELECT 
					*
				FROM
					proxy_list
				WHERE
					last_check <= ?
					order by last_check`
	//TODO do we need to filter out failed servers to lower the workload?
	_, e := db.Select(&list, query, time.Now().Add(
		-time.Duration(conf.Args.ProbeInterval)*time.Second).Format(util.DateTimeFormat))
	if e != nil {
		logrus.Errorln("failed to query stale proxy servers", e)
		return
	}
	logrus.Debugf("%d stale servers pending for health check", len(list))
	for _, p := range list {
		chjobs <- p
	}
}

func probe(chjobs <-chan *types.ProxyServer) {
	for i := 0; i < conf.Args.ProbeSize; i++ {
		go func() {
			for ps := range chjobs {
				status := types.FAIL
				if util.ValidateProxy(ps.Type, ps.Host, ps.Port) {
					status = types.OK
				}
				_, e := db.Exec(`update proxy_list set status = ?, last_check = ? where host = ? and port = ?`,
					status, util.Now(), ps.Host, ps.Port)
				if e != nil {
					logrus.Errorln("failed to update proxy server status", e)
				}
			}
		}()
	}
}
