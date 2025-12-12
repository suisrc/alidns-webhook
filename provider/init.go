package provider

import (
	"errors"

	"github.com/suisrc/webhook-dns/multi"
	"k8s.io/klog/v2"
)

var (
	ErrNoRecord = errors.New("record does not exist")
)

func init() {
	multi.DnsBuilders["custom"] = NewCustom
	multi.DnsBuilders["alidns"] = NewAlidns
	multi.DnsBuilders["dnspod"] = NewDnspod
	// multi.DnsBuilders["huawei"] = NewHuawei 没有账号，暂时没有测试

	klog.Info("Registered providers: custom, alidns, dnspod")
}
