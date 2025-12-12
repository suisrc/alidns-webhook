package provider

import (
	"github.com/suisrc/webhook-dns/multi"
)

func init() {
	multi.ClientBuilders["custom"] = NewCustom
	multi.ClientBuilders["alidns"] = NewAlidns
	multi.ClientBuilders["dnspod"] = NewDnspod
	multi.ClientBuilders["huawei"] = NewHuawei
}
