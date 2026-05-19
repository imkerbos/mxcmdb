package model

import "time"

// ProbeResult 探针采集结果
type ProbeResult struct {
	BaseModel
	AssetID         uint      `gorm:"index:idx_probe_result_asset_collected,priority:1;not null" json:"asset_id"`
	Hostname        string    `gorm:"size:128" json:"hostname"`
	CPU             string    `gorm:"type:text" json:"cpu"`
	Memory          string    `gorm:"type:text" json:"memory"`
	Disk            string    `gorm:"type:text" json:"disk"`
	Network         string    `gorm:"type:text" json:"network"`
	OS              string    `gorm:"size:128" json:"os"`
	Kernel          string    `gorm:"size:128" json:"kernel"`
	DockerVersion   string    `gorm:"size:64" json:"docker_version"`
	RunningServices string    `gorm:"type:text" json:"running_services"`
	SSHUsers        string    `gorm:"type:text" json:"ssh_users"`
	Uptime          string    `gorm:"size:64" json:"uptime"`
	DNS             string    `gorm:"type:text" json:"dns"`
	Gateway         string    `gorm:"size:256" json:"gateway"`
	Manufacturer    string    `gorm:"size:128" json:"manufacturer"`
	ProductModel    string    `gorm:"size:128" json:"product_model"`
	SerialNumber    string    `gorm:"size:128" json:"serial_number"`
	PublicIP        string    `gorm:"size:45" json:"public_ip"`
	Processes       string    `gorm:"type:text" json:"processes"`
	Listeners       string    `gorm:"type:text" json:"listeners"`
	CollectedAt     time.Time `gorm:"index:idx_probe_result_asset_collected,priority:2,sort:desc" json:"collected_at"`
	Status          string    `gorm:"size:32" json:"status"` // success / failed
	ErrorMessage    string    `gorm:"type:text" json:"error_message"`
}
