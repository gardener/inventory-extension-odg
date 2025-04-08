// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package types provides API models based on the upstream Open Delivery Gear
// API types for describing findings from Inventory.
//
// For more details about the upstream Open Delivery Gear models, please refer
// to [dso/model.py]
//
// [dso/model.py]: https://github.com/gardener/cc-utils/blob/master/dso/model.py
package types

import "time"

// Finding is a representation of the [InventoryFinding class]
//
// [InventoryFinding class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L622-L641
type Finding struct {
	// ProviderName specifies the name of the provider, from which orphan
	// resources originate from, e.g. AWS, Azure, GCP, OpenStack, etc.
	ProviderName string `json:"provider_name"`

	// ResourceKind specifies the kind of the orphan resource, e.g. Virtual
	// Machine, Public IP address, etc.
	ResourceKind string `json:"resource_kind"`

	// ResourceName specifies the unique name of the orphan resource in the
	// provider.
	ResourceName string `json:"resource_name"`

	// Summary specifies a short summary of the finding
	Summary string `json:"summary"`

	// Attributes specifies an optional set of attributes to associate with
	// the finding.
	Attributes map[string]string
}

// ArtefactKind is a representation of the upstream [ArtefactKind class]
//
// [ArtefactKind class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L183-L187
type ArtefactKind string

const (
	ArtefactKindArtefact ArtefactKind = "artefact"
	ArtefactKindResource ArtefactKind = "resource"
	ArtefactKindRuntime  ArtefactKind = "runtime"
	ArtefactKindSource   ArtefactKind = "source"
)

// Metadata is a representation of the upstream [Metadata class]
//
// [Metadata class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L306-L311
type Metadata struct {
	Datasource   string    `json:"datasource"`
	Type         string    `json:"type"`
	CreationDate time.Time `json:"creation_date"`
	LastUpdate   time.Time `json:"last_update"`
}

// LocalArtefactID is a representation of the upstream [LocalArtefactId class]
//
// [LocalArtefactId class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L140-L145
type LocalArtefactID struct {
	ArtefactName    string            `json:"artefact_name"`
	ArtefactType    string            `json:"artefact_type"`
	ArtefactVersion string            `json:"artefact_version"`
	ArtefactExtraID map[string]string `json:"artefact_extra_id"`
}

// ComponentArterfactID is a representation of the upstream
// [ComponentArtefactId class].
//
// [ComponentArtefactId class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L194-L200
type ComponentArterfactID struct {
	ComponentName    string          `json:"component_name"`
	ComponentVersion string          `json:"component_version"`
	Artefact         LocalArtefactID `json:"artefact"`
	ArtefactKind     ArtefactKind    `json:"artefact_kind"`
}

// ArtefactMetadata is a representation of the upstream [ArtefactMetadata class]
//
// [ArtefactMetadata class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L871-L906
type ArtefactMetadata struct {
	Artefact ComponentArterfactID `json:"artefact"`
	Meta     Metadata             `json:"meta"`
	Data     Finding              `json:"data"`
}
