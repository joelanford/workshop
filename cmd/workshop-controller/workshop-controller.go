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

package main

import (
	goflag "flag"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"k8s.io/apiserver/pkg/util/flag"
	"k8s.io/kubernetes/pkg/util/logs"

	"github.com/joelanford/workshop/cmd/workshop-controller/app"
	"github.com/joelanford/workshop/cmd/workshop-controller/app/options"
	"github.com/joelanford/workshop/pkg/workshop/controller/version"
)

func main() {
	config := options.NewWorkshopControllerConfig()
	config.AddFlags(pflag.CommandLine)
	flag.InitFlags()
	goflag.CommandLine.Parse([]string{})
	logs.InitLogs()
	defer logs.FlushLogs()

	version.PrintAndExitIfRequested()
	glog.V(0).Infof("version: %+v", version.VERSION)

	server := app.NewWorkshopControllerServerDefault(config)
	server.Run()
}
