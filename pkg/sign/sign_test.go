package sign

import (
	"context"
	"fmt"
	"github.com/pjoc-team/pay-gateway/pkg/configclient"
	"testing"
)

func TestSignConvert(t *testing.T) {
	s := "GBK 与 UTF-8 编码转换测试"
	gbk, err := Utf8ToGbk([]byte(s))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(gbk))
	}

	utf8, err := GbkToUtf8(gbk)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(utf8))
	}
}

func TestGenerateSign(t *testing.T) {
	config := &configclient.MerchantConfig{}
	config.GatewayRSAPrivateKey = `MIIEowIBAAKCAQEA2KaaJp7JeW91WlQCfZeS14US/ot9hIJViutv3JHojdgTx+8A8psStKaPl2Ac/MTJ/3mHeopCObmgjw/Au/Ne0PS1rveY0Pcazwnp+R1TDP2H9jagc3GJWS6cvHLB/B4uP3LOnPXN8ctwDVsF19b/howVKUKX6RAX7R2VAEyTIZJIEIQE0fNvRCWqbVv1RB3LU4cbQmW6nX8dP793fP8s/Lhzcj6vS6UKxLVl5CrCCGIJIBYc1mI8RbUYvGqwiONEnEwYvOioAoAlkMIXdFndIjngHe7JYfGW1NtPzHLG5yw8anYTD/3du7hJ/kSN0WM6NLa0P/vbR5+mxVdoRzY+kQIDAQABAoIBABaE2qkBADgbGbuV19xuENlN/7dtkFJhqbqS1kG6+M0llIjHkvWkoMEePvahCuJLIiPn4ekezdtqLAIyxPnERiq6BNh26+9sf+DdSvCV17gV8jfpXawiNQCME8aStw8Zo/z8VfWCpzFmz/LTbzwMIOs/TEPJpDiZb6M52+74BqMKfHTY14YOF8Xr4fiaUFpNTViHeOQXKzoG5PF4GLlhg7YNgEnjyc578izCoFp/xTjBBHQ7dtu+EnzmXD9QTlz7xUYt4P2TjUEBKy1oxSxDpgFL+BKYgRazilkrJ2hesbGCvbxDzcd4ivzpfmvqkN74Lq0vF9voL1JSd6D23l/R9bECgYEA7nbQgsK3ResReUMumJvE4y1sl2D+rt24QHlu+jOJqVkpAo0L4HLjvCX0Y8tBfG/hDc5iC12YILCn+EEb9bD2giURg7V+cA+K4IJrbLTbnna/UlvA1PFK3kHFosdCk5cRlpAppEBLQEUjlf7mjp6k2Xxy71ozg4KlB3wf3QCrs2MCgYEA6JUkiXd/lntdjb7V/QdwhVFdp/lzst0ClE4q04RNL8ZjwmSrYOAOGO2ktKOBG8lGT3P654/BASn9TMOXks8gPE3r/pN+21RGvOq2xtHNOrnV5g1RvlqHtwtv2RUxoEoTKPjBm6KDeLrPNCuGZ3bzYUUNAys66v3iWM5PK2s1GnsCgYEAjtKyx95/jmzgJlTKj7ScE8SdCX2ajHlXZaZVhZ1gkgFIwrJfrqqhI4tH+I1AR5tqm65EorIH72xe7h1w9ZJr0j8JYm1NsShd8WGrnYwlDZ/prxYtRFzQjpWuHXRit6r/acImbq3jZDcEvU3SIRF7gpc674iC2f1hgj4hh2hjbikCgYBWfG8ztv34xTMKrHYCOyv6R0FeXwJI9qoo39BJCx9wroMWHD0mLurPFj9y9IHkBTph/SzFwszwU97fFrRcYS0Jf6hL6Cj6AiKzyUviLs30EnqZq0ZEVIG27UfQH3NuuVzalXXZG9trn3vBWJYID1F9UCIAlai5DWOHxl/mM11x1QKBgCvi33kFLll6SVIKfkwt1Hja/DGlyq/M/xN4qn/wQwGKYzzIU+73SbQRg44GAiYVMJQrjISg/RVd4ClDxZ+A0cpumfpuSJdcT210L4u5FkuTQAmLZ2HOhTzKmW9/iR9koFHtTzTKhhYIgSWy9EWkQmcyrOKnEPYqMJjMobDJ1AuG`
	sign, err := GenerateSign(context.Background(), "utf-8", "草", config, SIGN_TYPE_SHA256_WITH_RSA)
	if err != nil {
		panic(err)
	}
	fmt.Println(sign)

}
