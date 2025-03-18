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
