// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package models

import (
	"net"
	"time"
)

// OrphanVirtualMachineAWS represents an AWS EC2 instance, which has been
// identified as being orphan.
type OrphanVirtualMachineAWS struct {
	Name         string    `bun:"name" json:"name"`
	Arch         string    `bun:"arch" json:"arch"`
	InstanceID   string    `bun:"instance_id" json:"instance_id"`
	InstanceType string    `bun:"instance_type" json:"instance_type"`
	State        string    `bun:"state" json:"state"`
	VpcID        string    `bun:"vpc_id" json:"vpc_id"`
	VpcName      string    `bun:"vpc_name" json:"vpc_name"`
	RegionName   string    `bun:"region_name" json:"region_name"`
	AccountID    string    `bun:"account_id" json:"account_id"`
	SubnetID     string    `bun:"subnet_id" json:"subnet_id"`
	Platform     string    `bun:"platform" json:"platform"`
	ImageID      string    `bun:"image_id" json:"image_id"`
	LaunchTime   time.Time `bun:"launch_time" json:"launch_time"`
}

// OrphanVirtualMachineGCP represents a GCP instance, which has been identified
// as being orphan.
type OrphanVirtualMachineGCP struct {
	Name                 string `bun:"name" json:"name"`
	Hostname             string `bun:"hostname" json:"hostname"`
	InstanceID           uint64 `bun:"instance_id" json:"instance_id"`
	ProjectID            string `bun:"project_id" json:"project_id"`
	Region               string `bun:"region" json:"region"`
	Zone                 string `bun:"zone" json:"zone"`
	CPUPlatform          string `bun:"cpu_platform" json:"cpu_platform"`
	Status               string `bun:"status" json:"status"`
	StatusMessage        string `bun:"status_message" json:"status_message"`
	CreationTimestamp    string `bun:"creation_timestamp" json:"creation_timestamp"`
	Description          string `bun:"description" json:"description"`
	LastStartTimestamp   string `bun:"last_start_timestamp" json:"last_start_timestamp"`
	LastStopTimestamp    string `bun:"last_stop_timestamp" json:"last_stop_timestamp"`
	LastSuspendTimestamp string `bun:"last_suspend_timestamp" json:"last_suspend_timestamp"`
	MachineType          string `bun:"machine_type" json:"machine_type"`
	GKEClusterName       string `bun:"gke_cluster_name" json:"gke_cluster_name"`
	GKEPoolName          string `bun:"gke_pool_name" json:"gke_pool_name"`
}

// OrphanVirtualMachineAzure represents an Azure virtual machine, which has been
// identified as being orphan.
type OrphanVirtualMachineAzure struct {
	Name                       string    `bun:"name" json:"name"`
	SubscriptionID             string    `bun:"subscription_id" json:"subscription_id"`
	ResourceGroup              string    `bun:"resource_group" json:"resource_group"`
	Location                   string    `bun:"location" json:"location"`
	ProvisioningState          string    `bun:"provisioning_state" json:"provisioning_state"`
	VirtualMachineCreatedAt    time.Time `bun:"vm_created_at" json:"vm_created_at"`
	VirtualMachineSize         string    `bun:"vm_size" json:"vm_size"`
	VirtualMachineAgentVersion string    `bun:"vm_agent_version" json:"vm_agent_version"`
	PowerState                 string    `bun:"power_state" json:"power_state"`
	HyperVGeneration           string    `bun:"hyper_v_gen" json:"hyper_v_gen"`
}

// OrphanVirtualMachineOpenStack represents an OpenStack server, which has been identified
// as being orphan.
type OrphanVirtualMachineOpenStack struct {
	ServerID         string `bun:"server_id" json:"server_id"`
	Name             string `bun:"name" json:"name"`
	ProjectID        string `bun:"project_id" json:"project_id"`
	ProjectName      string `bun:"project_name" json:"project_name"`
	Domain           string `bun:"domain" json:"domain"`
	Region           string `bun:"region" json:"region"`
	UserID           string `bun:"user_id" json:"user_id"`
	AvailabilityZone string `bun:"availability_zone" json:"availability_zone"`
	Status           string `bun:"status" json:"status"`
	ImageID          string `bun:"image_id" json:"image_id"`
	ServerCreatedAt  string `bun:"server_created_at" json:"server_created_at"`
	ServerUpdatedAt  string `bun:"server_updated_at" json:"server_updated_at"`
}

// OrphanPublicAddressGCP represents a GCP public IP address, which has been
// identified as being orphan.
type OrphanPublicAddressGCP struct {
	RuleID              uint64 `bun:"rule_id" json:"rule_id"`
	ProjectID           string `bun:"project_id" json:"project_id"`
	Name                string `bun:"name" json:"name"`
	IPAddress           net.IP `bun:"ip_address" json:"ip_address"`
	IPProtocol          string `bun:"ip_protocol" json:"ip_protocol"`
	IPVersion           string `bun:"ip_version" json:"ip_version"`
	AllPorts            bool   `bun:"all_ports" json:"all_ports"`
	AllowGlobalAccess   bool   `bun:"allow_global_access" json:"allow_global_access"`
	BackendService      string `bun:"backend_service" json:"backend_service"`
	CreationTimestamp   string `bun:"creation_timestamp" json:"creation_timestamp"`
	Description         string `bun:"description" json:"description"`
	LoadBalancingScheme string `bun:"load_balancing_scheme" json:"load_balancing_scheme"`
	Network             string `bun:"network" json:"network"`
	NetworkTier         string `bun:"network_tier" json:"network_tier"`
	PortRange           string `bun:"port_range" json:"port_range"`
	Region              string `bun:"region" json:"region"`
	ServiceLabel        string `bun:"service_label" json:"service_label"`
	ServiceName         string `bun:"service_name" json:"service_name"`
	Subnetwork          string `bun:"subnetwork" json:"subnetwork"`
	Target              string `bun:"target" json:"target"`
}
