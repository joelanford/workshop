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

type DeskStatusState string
type DeskStatusMessage string

const (
	DeskResourcePlural string = "desks"

	DefaultDeskVersion  string        = "latest"
	DefaultDeskLifespan time.Duration = time.Hour * 24 * 14

	DeskStatusStateInitializing DeskStatusState = "Initializing"
	DeskStatusStateReady        DeskStatusState = "Ready"
	DeskStatusStateExpired      DeskStatusState = "Expired"
	DeskStatusStateTerminating  DeskStatusState = "Terminating"

	DeskStatusMsgInitializing DeskStatusMessage = "Desk is initializing, not ready yet"
	DeskStatusMsgReady        DeskStatusMessage = "Desk is ready for use"
	DeskStatusMsgExpired      DeskStatusMessage = "Desk is expired and no longer accessible"
	DeskStatusMsgTerminating  DeskStatusMessage = "Desk is terminating"
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
	State   DeskStatusState   `json:"state,omitempty"`
	Message DeskStatusMessage `json:"message,omitempty"`
}

type Desk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              DeskSpec   `json:"spec"`
	Status            DeskStatus `json:"status,omitempty"`
}

func (d *Desk) DeepCopyObject() runtime.Object {
	return &Desk{
		TypeMeta:   d.TypeMeta,
		ObjectMeta: d.ObjectMeta,
		Spec:       d.Spec,
		Status:     d.Status,
	}
}

type DeskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Desk `json:"items"`
}

func (dl *DeskList) DeepCopyObject() runtime.Object {
	items := make([]Desk, len(dl.Items))
	copy(items, dl.Items)

	return &DeskList{
		TypeMeta: dl.TypeMeta,
		ListMeta: dl.ListMeta,
		Items:    items,
	}
}
