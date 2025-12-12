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
	multi.ClientBuilders["custom"] = NewCustom
	multi.ClientBuilders["alidns"] = NewAlidns
	multi.ClientBuilders["dnspod"] = NewDnspod
	// multi.ClientBuilders["huawei"] = NewHuawei 没有令牌，暂时没有测试

	klog.Info("Registered providers: custom, alidns, dnspod")
}
