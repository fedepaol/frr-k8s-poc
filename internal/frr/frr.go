// SPDX-License-Identifier:Apache-2.0

package frr

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	v1alpha1 "github.com/metallb/frrk8s/api/v1alpha1"
	"github.com/metallb/frrk8s/internal/logging"
)

// As the MetalLB controller should handle messages synchronously, there should
// no need to lock this data structure. TODO: confirm this.

type FRR struct {
	reloadConfig chan reloadEvent
	logLevel     string
	sync.Mutex
}

// Create a variable for os.Hostname() in order to make it easy to mock out
// in unit tests.
var osHostname = os.Hostname

func (f *FRR) ApplyConfig(k8sConfig v1alpha1.FRRConfiguration) error {
	hostname, err := osHostname()
	if err != nil {
		return err
	}

	config := &frrConfig{
		Hostname: hostname,
		Loglevel: f.logLevel,
		//BFDProfiles: sm.bfdProfiles,
		//ExtraConfig: sm.extraConfig,
	}

	for _, r := range k8sConfig.Spec.Routers {
		routerConfig, err := routerToFRRConfig(r)
		if err != nil {
			return err
		}
		config.Routers = append(config.Routers, routerConfig)
	}

	config.Loglevel = k8sConfig.Spec.LogLevel
	f.reloadConfig <- reloadEvent{config: config}
	return nil
}

var debounceTimeout = 3 * time.Second
var failureTimeout = time.Second * 5

func NewFRR(ctx context.Context, logLevel logging.Level) *FRR {
	logger, err := logging.Init(string(logLevel))
	if err != nil {
	}
	res := &FRR{
		reloadConfig: make(chan reloadEvent),
		logLevel:     logLevelToFRR(logLevel),
	}
	reload := func(config *frrConfig) error {
		return generateAndReloadConfigFile(config, logger)
	}

	debouncer(ctx, reload, res.reloadConfig, debounceTimeout, failureTimeout, logger)
	reloadValidator(ctx, logger, res.reloadConfig)
	return res
}

func reloadValidator(ctx context.Context, l log.Logger, reload chan<- reloadEvent) {
	var tickerIntervals = 30 * time.Second
	var prevReloadTimeStamp string

	ticker := time.NewTicker(tickerIntervals)
	go func() {
		select {
		case <-ticker.C:
			validateReload(l, &prevReloadTimeStamp, reload)
		case <-ctx.Done():
			return
		}
	}()
}

const statusFileName = "/etc/frr_reloader/.status"

func validateReload(l log.Logger, prevReloadTimeStamp *string, reload chan<- reloadEvent) {
	bytes, err := os.ReadFile(statusFileName)
	if err != nil {
		if !os.IsNotExist(err) {
			level.Error(l).Log("op", "reload-validate", "error", err, "cause", "readFile", "fileName", statusFileName)
		}
		return
	}

	lastReloadStatus := strings.Fields(string(bytes))
	if len(lastReloadStatus) != 2 {
		level.Error(l).Log("op", "reload-validate", "error", err, "cause", "Fields", "bytes", string(bytes))
		return
	}

	timeStamp, status := lastReloadStatus[0], lastReloadStatus[1]
	if timeStamp == *prevReloadTimeStamp {
		return
	}

	*prevReloadTimeStamp = timeStamp

	if strings.Compare(status, "failure") == 0 {
		level.Error(l).Log("op", "reload-validate", "error", fmt.Errorf("reload failure"),
			"cause", "frr reload failed", "status", status)
		reload <- reloadEvent{useOld: true}
		return
	}

	level.Info(l).Log("op", "reload-validate", "success", "reloaded config")
}

func logLevelToFRR(level logging.Level) string {
	// Allowed frr log levels are: emergencies, alerts, critical,
	// 		errors, warnings, notifications, informational, or debugging
	switch level {
	case logging.LevelAll, logging.LevelDebug:
		return "debugging"
	case logging.LevelInfo:
		return "informational"
	case logging.LevelWarn:
		return "warnings"
	case logging.LevelError:
		return "error"
	case logging.LevelNone:
		return "emergencies"
	}

	return "informational"
}

func sortMap[T any](toSort map[string]T) []T {
	keys := make([]string, 0)
	for k := range toSort {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	res := make([]T, 0)
	for _, k := range keys {
		res = append(res, toSort[k])
	}
	return res
}
