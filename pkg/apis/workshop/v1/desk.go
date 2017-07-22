/*
Copyright 2017 The Kubernetes Authors.

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

package v1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type DeskState string

const (
	DeskKind           string = "Desk"
	DeskResourcePlural string = "desks"
	DeskCRDName        string = DeskResourcePlural + "." + GroupName

	DeskDefaultVersion string        = "latest"
	DeskMaxLifespan    time.Duration = time.Hour * 24 * 14

	DeskStateInitializing DeskState = "Initializing"
	DeskStateReady        DeskState = "Ready"
	DeskStateExpired      DeskState = "Expired"
	DeskStateTerminating  DeskState = "Terminating"
)

type DeskSpec struct {
	// Version of the desk to be deployed. (optional; default "latest")
	Version string `json:"version,omitempty"`

	// Owner of the desk (required)
	Owner string `json:"owner"`

	// Time after which desk will be auto-deleted. (optional; default - 2 weeks after creation)
	ExpirationTimestamp metav1.Time `json:"expirationTimestamp,omitempty"`
}

type DeskStatus struct {
	State DeskState `json:"state,omitempty"`
}

type Desk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DeskSpec   `json:"spec"`
	Status            DeskStatus `json:"status,omitempty"`
}

func (d *Desk) DeepCopyObject() runtime.Object {
	dCopy := *d
	return &dCopy
}

type DeskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Desk `json:"items"`
}

func (dl *DeskList) DeepCopyObject() runtime.Object {
	dlCopy := *dl

	items := make([]Desk, len(dl.Items))
	copy(items, dl.Items)
	dlCopy.Items = items

	return &dlCopy
}
