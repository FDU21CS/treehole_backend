package utils

import (
	"testing"
	"treehole_backend/config"
)

func TestSendEmail(t *testing.T) {
	config.InitConfig()

	err := SendCodeEmail("123456", "21307130001@m.fudan.edu.cn")
	if err != nil {
		t.Fatal(err)
	}
}
