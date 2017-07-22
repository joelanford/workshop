/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package options contains flags for initializing a proxy.
package options

import (
	"fmt"
	_ "net/http/pprof"
	"net/url"
	"os"
	"time"

	"strings"

	"github.com/spf13/pflag"
)

type WorkshopControllerConfig struct {
	KubeConfigFile string
	KubeMasterURL  string

	Clean bool

	DesksDomain string

	HealthzPort        int
	InitialSyncTimeout time.Duration
}

func NewWorkshopControllerConfig() *WorkshopControllerConfig {
	return &WorkshopControllerConfig{
		DesksDomain:        "desks.workshop.lanford.io",
		HealthzPort:        8081,
		InitialSyncTimeout: 60 * time.Second,
	}
}

type kubeMasterURLVar struct {
	val *string
}

func (m kubeMasterURLVar) Set(v string) error {
	parsedURL, err := url.Parse(os.ExpandEnv(v))
	if err != nil {
		return fmt.Errorf("failed to parse kube-master-url")
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" || parsedURL.Host == ":" {
		return fmt.Errorf("invalid kube-master-url specified")
	}
	*m.val = v
	return nil
}

func (m kubeMasterURLVar) String() string {
	return *m.val
}

func (m kubeMasterURLVar) Type() string {
	return "string"
}

type desksDomainVar struct {
	val *string
}

func (m desksDomainVar) Set(v string) error {
	parsedURL, err := url.Parse(v)
	if err != nil {
		return fmt.Errorf("failed to parse desks-domain")
	}
	if parsedURL.Scheme != "" || parsedURL.Path != "" || strings.Contains(parsedURL.Host, ":") {
		return fmt.Errorf("invalid desks-domain specified")
	}
	*m.val = v
	return nil
}

func (m desksDomainVar) String() string {
	return *m.val
}

func (m desksDomainVar) Type() string {
	return "string"
}

func (s *WorkshopControllerConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.KubeConfigFile, "kubecfg-file", s.KubeConfigFile,
		"Location of kubecfg file for access to kubernetes master service;"+
			" --kube-master-url overrides the URL part of this; if this is not"+
			" provided, defaults to service account tokens")
	fs.Var(kubeMasterURLVar{&s.KubeMasterURL}, "kube-master-url",
		"URL to reach kubernetes master. Env variables in this flag will be expanded.")

	fs.BoolVar(&s.Clean, "clean", s.Clean,
		"Clean desk-controller resources and exit.")

	fs.IntVar(&s.HealthzPort, "healthz-port", s.HealthzPort,
		"port on which to serve a kube-dns HTTP readiness probe.")

	fs.DurationVar(&s.InitialSyncTimeout, "initial-sync-timeout", s.InitialSyncTimeout,
		"Timeout for initial resource sync.")

	fs.Var(desksDomainVar{&s.DesksDomain}, "desks-domain",
		"DNS domain at which desks will be reachable (e.g. \"<deskName>.<domain>\").")
}
