//go:build integration_tests || unit_tests || domain_tests || domain_unit_tests

package domain

import (
	"testing"
)

func TestBelongsToISP(t *testing.T) {
	ipinfo := IPInfo{IP: "1.1.1.1", OrgName: "DIGI"}

	trueResult := ipinfo.BelongsToISP("DIGI")
	falseResult := ipinfo.BelongsToISP("TELEFONICA")

	if trueResult == false {
		t.Errorf("IP info with DIGI OrgName should return true when BelongsToISP is called ith \"DIGI\" ISP value")
	}
	if falseResult == true {
		t.Errorf("IP info with DIGI OrgName should return false when BelongsToISP is called ith \"TELEFONICA\" ISP value")
	}

}
