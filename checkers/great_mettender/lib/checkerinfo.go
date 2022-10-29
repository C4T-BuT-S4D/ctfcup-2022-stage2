package lib

type CheckerInfo struct {
	Vulns      int  `json:"vulns"`
	Timeout    int  `json:"timeout"`
	AttackData bool `json:"attack_data"`
}
