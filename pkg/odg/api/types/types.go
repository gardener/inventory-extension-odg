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

import (
	"time"

	"cloud.google.com/go/civil"
)

// SeverityLevel specifies the level of severity for a finding.
type SeverityLevel string

const (
	// SeverityLevelLow specifies a finding with low severity level
	SeverityLevelLow = "LOW"

	// SeverityLevelMedium specifies a finding with medium severity level
	SeverityLevelMedium = "MEDIUM"

	// SeverityLevelHigh specifies a finding with high severity level
	SeverityLevelHigh = "HIGH"
)

// ArtefactKind is a representation of the upstream [ArtefactKind class]
//
// [ArtefactKind class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L183-L187
type ArtefactKind string

// The various artefact kinds supported by the upstream ODG API service.
const (
	ArtefactKindArtefact ArtefactKind = "artefact"
	ArtefactKindResource ArtefactKind = "resource"
	ArtefactKindRuntime  ArtefactKind = "runtime"
	ArtefactKindSource   ArtefactKind = "source"
)

// Datasource is a representation of the upstream [Datasource class].
//
// [Datasource class]: https://github.com/dnaeon/cc-utils/blob/5df6327a17b9358f772084124f997d26b0fdf4ea/dso/model.py#L59-L70
type Datasource string

const (
	// DatasourceInventory is the Inventory datasource for findings
	DatasourceInventory Datasource = "inventory"
)

// Datatype is a representation of the upstream [Datatype class].
//
// [Datatype class]: https://github.com/dnaeon/cc-utils/blob/5df6327a17b9358f772084124f997d26b0fdf4ea/dso/model.py#L270-L286
type Datatype string

const (
	// DatatypeInventory represents a finding from the Inventory system
	DatatypeInventory Datatype = "finding/inventory"

	// DatatypeArtefactScanInfo is a meta artefact, which represents that a
	// scan from given datasource has been performed.
	DatatypeArtefactScanInfo = "meta/artefact_scan_info"
)

// ResourceKind represents the kind of orphan resource, which will be submitted
// to the Delivery Service.
type ResourceKind string

const (
	// ResourceKindVirtualMachineAWS represents an AWS Virtual Machine
	// resource.
	ResourceKindVirtualMachineAWS ResourceKind = "aws/virtual-machine"

	// ResourceKindVirtualMachineGCP represents a GCP Virtual Machine
	// resource.
	ResourceKindVirtualMachineGCP ResourceKind = "gcp/virtual-machine"

	// ResourceKindVirtualMachineAzure represents an Azure Virtual Machine
	// resource.
	ResourceKindVirtualMachineAzure ResourceKind = "az/virtual-machine"

	// ResourceKindVirtualMachineOpenStack represents a OpenStack Virtual Machine
	// resource.
	ResourceKindVirtualMachineOpenStack ResourceKind = "openstack/virtual-machine"

	// ResourceKindIPAddressGCP represents a GCP Public IP address resource.
	ResourceKindIPAddressGCP ResourceKind = "gcp/public-ip-address"
)

// ProviderName specifies the name of the provider, from which orphan resources
// originate from.
type ProviderName string

const (
	// ProviderNameAWS represents AWS as the origin of orphan resources.
	ProviderNameAWS ProviderName = "aws"

	// ProviderNameGCP represents GCP as the origin of orphan resources.
	ProviderNameGCP ProviderName = "gcp"

	// ProviderNameAzure represents Azure as the origin of orphan resources.
	ProviderNameAzure ProviderName = "azure"

	// ProviderNameOpenStack represents OpenStack as the origin of orphan resources.
	ProviderNameOpenStack ProviderName = "openstack"
)

// Finding is a representation of the [InventoryFinding class]
//
// [InventoryFinding class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L622-L641
type Finding struct {
	// Severity specifies the severity of the finding
	Severity SeverityLevel `json:"severity"`

	// ProviderName specifies the name of the provider, from which orphan
	// resources originate from, e.g. AWS, Azure, GCP, OpenStack, etc.
	ProviderName ProviderName `json:"provider_name"`

	// ResourceKind specifies the kind of the orphan resource, e.g. Virtual
	// Machine, Public IP address, etc.
	ResourceKind ResourceKind `json:"resource_kind"`

	// ResourceName specifies the unique name of the orphan resource in the
	// provider.
	ResourceName string `json:"resource_name"`

	// Summary specifies a short summary of the finding
	Summary string `json:"summary"`

	// Attributes specifies an optional set of attributes to associate with
	// the finding.
	Attributes any `json:"attributes"`
}

// Metadata is a representation of the upstream [Metadata class]
//
// [Metadata class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L306-L311
type Metadata struct {
	Datasource   Datasource `json:"datasource"`
	Type         Datatype   `json:"type"`
	CreationDate time.Time  `json:"creation_date"`
	LastUpdate   time.Time  `json:"last_update"`
}

// LocalArtefactID is a representation of the upstream [LocalArtefactId class]
//
// [LocalArtefactId class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L140-L145
type LocalArtefactID struct {
	ArtefactName    string            `json:"artefact_name,omitempty"`
	ArtefactType    string            `json:"artefact_type,omitempty"`
	ArtefactVersion string            `json:"artefact_version,omitempty"`
	ArtefactExtraID map[string]string `json:"artefact_extra_id,omitempty"`
}

// ComponentArtefactID is a representation of the upstream
// [ComponentArtefactId class].
//
// [ComponentArtefactId class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L194-L200
type ComponentArtefactID struct {
	ComponentName    string          `json:"component_name,omitempty"`
	ComponentVersion string          `json:"component_version,omitempty"`
	Artefact         LocalArtefactID `json:"artefact"`
	ArtefactKind     ArtefactKind    `json:"artefact_kind,omitempty"`
}

// ArtefactMetadata is a representation of the upstream [ArtefactMetadata class]
//
// [ArtefactMetadata class]: https://github.com/gardener/cc-utils/blob/af54ca4f80b6b96dbb981d7c9ea080239f552a49/dso/model.py#L871-L906
type ArtefactMetadata struct {
	Artefact      ComponentArtefactID `json:"artefact"`
	Meta          Metadata            `json:"meta"`
	Data          Finding             `json:"data"`
	DiscoveryDate civil.Date          `json:"discovery_date"`
}

// ArtefactMetadataGroup represents a group of [ArtefactMetadata] items.
type ArtefactMetadataGroup struct {
	// Entries contains the group of [ArtefactMetadata] items.
	Entries []ArtefactMetadata `json:"entries"`
}

// ComponentArtefactIDGroup represents a group of [ComponentArtefactID] items.
type ComponentArtefactIDGroup struct {
	// Entries contains the group of [ComponentArtefactID] items.
	Entries []ComponentArtefactID `json:"entries"`
}

// RuntimeArtefactGroup represents a group of [ComponentArtefactID] items, which
// are submitted to the Delivery Service API as `runtime-artefacts'.
type RuntimeArtefactGroup struct {
	Artefacts []ComponentArtefactID `json:"artefacts"`
}

// RuntimeArtefactMetadata represents the metadata for a runtime artefact as
// retrieved from the remote Delivery Service API.
type RuntimeArtefactMetadata struct {
	Name              string            `json:"name,omitempty"`
	UID               string            `json:"uid,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
}

// RuntimeArtefactSpec represents the spec for a runtime artefact as retrieved
// from the remote Delivery Service API.
type RuntimeArtefactSpec struct {
	CreationData time.Time           `json:"creation_date"`
	Artefact     ComponentArtefactID `json:"artefact"`
}

// RuntimeArtefactResultItem represents a runtime artefact retrieved from the
// remote Delivery Service API.
type RuntimeArtefactResultItem struct {
	Metadata RuntimeArtefactMetadata `json:"metadata"`
	Spec     RuntimeArtefactSpec     `json:"spec"`
}
