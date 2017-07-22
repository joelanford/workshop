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

package v1

import "github.com/joelanford/workshop/pkg/apis/workshop/v1"

// The DeskExpansion interface allows manually adding extra methods to the DeskInterface.
type DeskExpansion interface {
	Finalize(item *v1.Desk) (*v1.Desk, error)
}

// Finalize takes the representation of a desk to update.  Returns the server's representation of the desk, and an error, if it occurs.
func (c *desks) Finalize(desk *v1.Desk) (result *v1.Desk, err error) {
	result = &v1.Desk{}
	err = c.client.Put().Resource("desks").Name(desk.Name).SubResource("finalize").Body(desk).Do().Into(result)
	return
}
