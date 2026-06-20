package domain

type IPInfo struct {
	IP      string
	OrgName string
}

func (ipinfo IPInfo) BelongsToISP(isp string) bool {
	return ipinfo.OrgName == isp
}
