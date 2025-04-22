// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	odgapi "github.com/gardener/inventory-extension-odg/pkg/odg/api/client"
)

// Client is the Open Delivery Gear API client used by the various tasks.
var Client *odgapi.Client

// SetClient sets the default API client to the given [odgapi.Client]
func SetClient(c *odgapi.Client) {
	Client = c
}
