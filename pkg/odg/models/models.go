// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package models

import "time"

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
