package models

type Deployments []struct{
	Name 		string		`json:"name,omitempty"`
	Releases 	[]Release	`json:"releases,omitempty"`
	Stemcells 	[]Stemcell	`json:"stemcells,omitempty"`
}

type Release struct {
	Name 	string		`json:"name,omitempty"`
	Version string		`json:"version,omitempty"`
}

type Stemcell struct {
	Name 	string		`json:"name,omitempty"`
	Version string		`json:"version,omitempty"`
}

type Vitals struct {
	Job 		string 		`json:"job,omitemtpy"`
	Index 		string 		`json:"index,omitemtpy"`
	State 		string		`json:"state,omitemtpy"`
	Az		string 		`json:"az,omitemtpy"`
	Type 		string 		`json:"type,omitemtpy"`
	Ip 		string 		`json:"ip,omitemtpy"`
	Load_avg1 	float64 	`json:"load_avg1,omitemtpy"`
	Load_avg5 	float64 	`json:"load_avg5,omitemtpy"`
	Load_avg15 	float64 	`json:"load_avg15,omitemtpy"`
	Cpu_user 	float64 	`json:"cpu_user,omitemtpy"`
	Cpu_sys 	float64 	`json:"cpu_sys,omitemtpy"`
	Cpu_wait 	float64 	`json:"cpu_wait,omitemtpy"`
	Mem_usage 	float64 	`json:"mem_usage,omitemtpy"`
	Disk_usage 	float64 	`json:"disk_usage,omitemtpy"`
	Ephem_usage 	float64 	`json:"ephemeral_usage,omitemtpy"`
	Persist_usage 	float64 	`json:"persistent_usage,omitemtpy"`
}