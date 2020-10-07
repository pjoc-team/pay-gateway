package sign

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"github.com/blademainer/commons/pkg/util"
	"testing"
)

var privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2KaaJp7JeW91WlQCfZeS14US/ot9hIJViutv3JHojdgTx+8A
8psStKaPl2Ac/MTJ/3mHeopCObmgjw/Au/Ne0PS1rveY0Pcazwnp+R1TDP2H9jag
c3GJWS6cvHLB/B4uP3LOnPXN8ctwDVsF19b/howVKUKX6RAX7R2VAEyTIZJIEIQE
0fNvRCWqbVv1RB3LU4cbQmW6nX8dP793fP8s/Lhzcj6vS6UKxLVl5CrCCGIJIBYc
1mI8RbUYvGqwiONEnEwYvOioAoAlkMIXdFndIjngHe7JYfGW1NtPzHLG5yw8anYT
D/3du7hJ/kSN0WM6NLa0P/vbR5+mxVdoRzY+kQIDAQABAoIBABaE2qkBADgbGbuV
19xuENlN/7dtkFJhqbqS1kG6+M0llIjHkvWkoMEePvahCuJLIiPn4ekezdtqLAIy
xPnERiq6BNh26+9sf+DdSvCV17gV8jfpXawiNQCME8aStw8Zo/z8VfWCpzFmz/LT
bzwMIOs/TEPJpDiZb6M52+74BqMKfHTY14YOF8Xr4fiaUFpNTViHeOQXKzoG5PF4
GLlhg7YNgEnjyc578izCoFp/xTjBBHQ7dtu+EnzmXD9QTlz7xUYt4P2TjUEBKy1o
xSxDpgFL+BKYgRazilkrJ2hesbGCvbxDzcd4ivzpfmvqkN74Lq0vF9voL1JSd6D2
3l/R9bECgYEA7nbQgsK3ResReUMumJvE4y1sl2D+rt24QHlu+jOJqVkpAo0L4HLj
vCX0Y8tBfG/hDc5iC12YILCn+EEb9bD2giURg7V+cA+K4IJrbLTbnna/UlvA1PFK
3kHFosdCk5cRlpAppEBLQEUjlf7mjp6k2Xxy71ozg4KlB3wf3QCrs2MCgYEA6JUk
iXd/lntdjb7V/QdwhVFdp/lzst0ClE4q04RNL8ZjwmSrYOAOGO2ktKOBG8lGT3P6
54/BASn9TMOXks8gPE3r/pN+21RGvOq2xtHNOrnV5g1RvlqHtwtv2RUxoEoTKPjB
m6KDeLrPNCuGZ3bzYUUNAys66v3iWM5PK2s1GnsCgYEAjtKyx95/jmzgJlTKj7Sc
E8SdCX2ajHlXZaZVhZ1gkgFIwrJfrqqhI4tH+I1AR5tqm65EorIH72xe7h1w9ZJr
0j8JYm1NsShd8WGrnYwlDZ/prxYtRFzQjpWuHXRit6r/acImbq3jZDcEvU3SIRF7
gpc674iC2f1hgj4hh2hjbikCgYBWfG8ztv34xTMKrHYCOyv6R0FeXwJI9qoo39BJ
Cx9wroMWHD0mLurPFj9y9IHkBTph/SzFwszwU97fFrRcYS0Jf6hL6Cj6AiKzyUvi
Ls30EnqZq0ZEVIG27UfQH3NuuVzalXXZG9trn3vBWJYID1F9UCIAlai5DWOHxl/m
M11x1QKBgCvi33kFLll6SVIKfkwt1Hja/DGlyq/M/xN4qn/wQwGKYzzIU+73SbQR
g44GAiYVMJQrjISg/RVd4ClDxZ+A0cpumfpuSJdcT210L4u5FkuTQAmLZ2HOhTzK
mW9/iR9koFHtTzTKhhYIgSWy9EWkQmcyrOKnEPYqMJjMobDJ1AuG
-----END RSA PRIVATE KEY-----`

var publicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2KaaJp7JeW91WlQCfZeS
14US/ot9hIJViutv3JHojdgTx+8A8psStKaPl2Ac/MTJ/3mHeopCObmgjw/Au/Ne
0PS1rveY0Pcazwnp+R1TDP2H9jagc3GJWS6cvHLB/B4uP3LOnPXN8ctwDVsF19b/
howVKUKX6RAX7R2VAEyTIZJIEIQE0fNvRCWqbVv1RB3LU4cbQmW6nX8dP793fP8s
/Lhzcj6vS6UKxLVl5CrCCGIJIBYc1mI8RbUYvGqwiONEnEwYvOioAoAlkMIXdFnd
IjngHe7JYfGW1NtPzHLG5yw8anYTD/3du7hJ/kSN0WM6NLa0P/vbR5+mxVdoRzY+
kQIDAQAB
-----END PUBLIC KEY-----`

func TestSignPKCS1v15(t *testing.T) {
	s := util.RandString(10240)
	sign, _ := SignPKCS1v15([]byte(s), []byte(privateKey), crypto.SHA256)
	fmt.Println(base64.StdEncoding.EncodeToString(sign))
	if err := VerifyPKCS1v15([]byte(s), sign, []byte(publicKey), crypto.SHA256); err != nil {
		fmt.Println("Check sign error: ", err.Error())
	} else {
		fmt.Println("Check sign success!")
	}
}
func TestSignPKCS1v15WithStringKey(t *testing.T) {
	s := util.RandString(10240)
	privateKeyStr := `MIIEowIBAAKCAQEA2KaaJp7JeW91WlQCfZeS14US/ot9hIJViutv3JHojdgTx+8A8psStKaPl2Ac/MTJ/3mHeopCObmgjw/Au/Ne0PS1rveY0Pcazwnp+R1TDP2H9jagc3GJWS6cvHLB/B4uP3LOnPXN8ctwDVsF19b/howVKUKX6RAX7R2VAEyTIZJIEIQE0fNvRCWqbVv1RB3LU4cbQmW6nX8dP793fP8s/Lhzcj6vS6UKxLVl5CrCCGIJIBYc1mI8RbUYvGqwiONEnEwYvOioAoAlkMIXdFndIjngHe7JYfGW1NtPzHLG5yw8anYTD/3du7hJ/kSN0WM6NLa0P/vbR5+mxVdoRzY+kQIDAQABAoIBABaE2qkBADgbGbuV19xuENlN/7dtkFJhqbqS1kG6+M0llIjHkvWkoMEePvahCuJLIiPn4ekezdtqLAIyxPnERiq6BNh26+9sf+DdSvCV17gV8jfpXawiNQCME8aStw8Zo/z8VfWCpzFmz/LTbzwMIOs/TEPJpDiZb6M52+74BqMKfHTY14YOF8Xr4fiaUFpNTViHeOQXKzoG5PF4GLlhg7YNgEnjyc578izCoFp/xTjBBHQ7dtu+EnzmXD9QTlz7xUYt4P2TjUEBKy1oxSxDpgFL+BKYgRazilkrJ2hesbGCvbxDzcd4ivzpfmvqkN74Lq0vF9voL1JSd6D23l/R9bECgYEA7nbQgsK3ResReUMumJvE4y1sl2D+rt24QHlu+jOJqVkpAo0L4HLjvCX0Y8tBfG/hDc5iC12YILCn+EEb9bD2giURg7V+cA+K4IJrbLTbnna/UlvA1PFK3kHFosdCk5cRlpAppEBLQEUjlf7mjp6k2Xxy71ozg4KlB3wf3QCrs2MCgYEA6JUkiXd/lntdjb7V/QdwhVFdp/lzst0ClE4q04RNL8ZjwmSrYOAOGO2ktKOBG8lGT3P654/BASn9TMOXks8gPE3r/pN+21RGvOq2xtHNOrnV5g1RvlqHtwtv2RUxoEoTKPjBm6KDeLrPNCuGZ3bzYUUNAys66v3iWM5PK2s1GnsCgYEAjtKyx95/jmzgJlTKj7ScE8SdCX2ajHlXZaZVhZ1gkgFIwrJfrqqhI4tH+I1AR5tqm65EorIH72xe7h1w9ZJr0j8JYm1NsShd8WGrnYwlDZ/prxYtRFzQjpWuHXRit6r/acImbq3jZDcEvU3SIRF7gpc674iC2f1hgj4hh2hjbikCgYBWfG8ztv34xTMKrHYCOyv6R0FeXwJI9qoo39BJCx9wroMWHD0mLurPFj9y9IHkBTph/SzFwszwU97fFrRcYS0Jf6hL6Cj6AiKzyUviLs30EnqZq0ZEVIG27UfQH3NuuVzalXXZG9trn3vBWJYID1F9UCIAlai5DWOHxl/mM11x1QKBgCvi33kFLll6SVIKfkwt1Hja/DGlyq/M/xN4qn/wQwGKYzzIU+73SbQRg44GAiYVMJQrjISg/RVd4ClDxZ+A0cpumfpuSJdcT210L4u5FkuTQAmLZ2HOhTzKmW9/iR9koFHtTzTKhhYIgSWy9EWkQmcyrOKnEPYqMJjMobDJ1AuG`
	sign, _ := SignPKCS1v15WithStringKey([]byte(s), privateKeyStr, crypto.SHA256)
	fmt.Println(base64.StdEncoding.EncodeToString(sign))
	if err := VerifyPKCS1v15WithStringKey([]byte(s), sign, publicKey, crypto.SHA256); err != nil {
		fmt.Println("Check sign error: ", err.Error())
	} else {
		fmt.Println("Check sign success!")
	}
}

func TestSignMerchant(t *testing.T) {
	message := "app_id=1&channel_id=demo&method=test&order_time=2018-10-28 12:00:00&out_trade_no=201810281512542133254&pay_amount=1&sign_type=RSA"
	privateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA2KaaJp7JeW91WlQCfZeS14US/ot9hIJViutv3JHojdgTx+8A
8psStKaPl2Ac/MTJ/3mHeopCObmgjw/Au/Ne0PS1rveY0Pcazwnp+R1TDP2H9jag
c3GJWS6cvHLB/B4uP3LOnPXN8ctwDVsF19b/howVKUKX6RAX7R2VAEyTIZJIEIQE
0fNvRCWqbVv1RB3LU4cbQmW6nX8dP793fP8s/Lhzcj6vS6UKxLVl5CrCCGIJIBYc
1mI8RbUYvGqwiONEnEwYvOioAoAlkMIXdFndIjngHe7JYfGW1NtPzHLG5yw8anYT
D/3du7hJ/kSN0WM6NLa0P/vbR5+mxVdoRzY+kQIDAQABAoIBABaE2qkBADgbGbuV
19xuENlN/7dtkFJhqbqS1kG6+M0llIjHkvWkoMEePvahCuJLIiPn4ekezdtqLAIy
xPnERiq6BNh26+9sf+DdSvCV17gV8jfpXawiNQCME8aStw8Zo/z8VfWCpzFmz/LT
bzwMIOs/TEPJpDiZb6M52+74BqMKfHTY14YOF8Xr4fiaUFpNTViHeOQXKzoG5PF4
GLlhg7YNgEnjyc578izCoFp/xTjBBHQ7dtu+EnzmXD9QTlz7xUYt4P2TjUEBKy1o
xSxDpgFL+BKYgRazilkrJ2hesbGCvbxDzcd4ivzpfmvqkN74Lq0vF9voL1JSd6D2
3l/R9bECgYEA7nbQgsK3ResReUMumJvE4y1sl2D+rt24QHlu+jOJqVkpAo0L4HLj
vCX0Y8tBfG/hDc5iC12YILCn+EEb9bD2giURg7V+cA+K4IJrbLTbnna/UlvA1PFK
3kHFosdCk5cRlpAppEBLQEUjlf7mjp6k2Xxy71ozg4KlB3wf3QCrs2MCgYEA6JUk
iXd/lntdjb7V/QdwhVFdp/lzst0ClE4q04RNL8ZjwmSrYOAOGO2ktKOBG8lGT3P6
54/BASn9TMOXks8gPE3r/pN+21RGvOq2xtHNOrnV5g1RvlqHtwtv2RUxoEoTKPjB
m6KDeLrPNCuGZ3bzYUUNAys66v3iWM5PK2s1GnsCgYEAjtKyx95/jmzgJlTKj7Sc
E8SdCX2ajHlXZaZVhZ1gkgFIwrJfrqqhI4tH+I1AR5tqm65EorIH72xe7h1w9ZJr
0j8JYm1NsShd8WGrnYwlDZ/prxYtRFzQjpWuHXRit6r/acImbq3jZDcEvU3SIRF7
gpc674iC2f1hgj4hh2hjbikCgYBWfG8ztv34xTMKrHYCOyv6R0FeXwJI9qoo39BJ
Cx9wroMWHD0mLurPFj9y9IHkBTph/SzFwszwU97fFrRcYS0Jf6hL6Cj6AiKzyUvi
Ls30EnqZq0ZEVIG27UfQH3NuuVzalXXZG9trn3vBWJYID1F9UCIAlai5DWOHxl/m
M11x1QKBgCvi33kFLll6SVIKfkwt1Hja/DGlyq/M/xN4qn/wQwGKYzzIU+73SbQR
g44GAiYVMJQrjISg/RVd4ClDxZ+A0cpumfpuSJdcT210L4u5FkuTQAmLZ2HOhTzK
mW9/iR9koFHtTzTKhhYIgSWy9EWkQmcyrOKnEPYqMJjMobDJ1AuG
-----END RSA PRIVATE KEY-----`
	publicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2KaaJp7JeW91WlQCfZeS
14US/ot9hIJViutv3JHojdgTx+8A8psStKaPl2Ac/MTJ/3mHeopCObmgjw/Au/Ne
0PS1rveY0Pcazwnp+R1TDP2H9jagc3GJWS6cvHLB/B4uP3LOnPXN8ctwDVsF19b/
howVKUKX6RAX7R2VAEyTIZJIEIQE0fNvRCWqbVv1RB3LU4cbQmW6nX8dP793fP8s
/Lhzcj6vS6UKxLVl5CrCCGIJIBYc1mI8RbUYvGqwiONEnEwYvOioAoAlkMIXdFnd
IjngHe7JYfGW1NtPzHLG5yw8anYTD/3du7hJ/kSN0WM6NLa0P/vbR5+mxVdoRzY+
kQIDAQAB
-----END PUBLIC KEY-----`

	sign, err := SignPKCS1v15([]byte(message), []byte(privateKey), crypto.SHA256)
	if err != nil {
		fmt.Println("sign error: ", err.Error())
		return
	}
	signBase64 := string(base64.StdEncoding.EncodeToString(sign))
	fmt.Println("signBase64: ", signBase64)
	bytes, e := base64.StdEncoding.DecodeString(signBase64)
	e = VerifyPKCS1v15([]byte(message), bytes, []byte(publicKey), crypto.SHA256)
	if e != nil {
		fmt.Println("verify error: ", e.Error())
	}

	//source := "app_id=1&channel_id=demo&method=test&order_time=2018-10-28 12:00:00&pay_amount=1&sign_type=RSA"
	////signString := "OvsyKwWYeQdofkM0B9KwNpyajNBsd+yxle6EJzLxIE/2IrBICDwt77ydkd7NbYCsE9wfMH5ctNGCRfNhb58iqxZs+wjWz5t7KQNsJrQAWyz7lN8m0aHF7HcnH6XBdbjZ1FaaV9JSKF5DpVptneLD7jB5EsEaUC2b4u9hdc29j/HtLUWAaXNupxonM1DVZGn5JybhLZiWhLxBqzfDbkfshDTUdL0ZJdHWsn1FVFfWbAiAim8Sk4eLrhFP9NbnTL7GOgVgTnURlzZVeEhNd9vBT1cY2Gp51Nn2QREam0heb4UKG9n6fxAUxPCo3avQUhjvVSfxcKn8wz7qSHCNOYCquQ=="
	//signString := "OvsyKwWYeQdofkM0B9KwNpyajNBsd+yxle6EJzLxIE/2IrBICDwt77ydkd7NbYCsE9wfMH5ctNGCRfNhb58iqxZs+wjWz5t7KQNsJrQAWyz7lN8m0aHF7HcnH6XBdbjZ1FaaV9JSKF5DpVptneLD7jB5EsEaUC2b4u9hdc29j/HtLUWAaXNupxonM1DVZGn5JybhLZiWhLxBqzfDbkfshDTUdL0ZJdHWsn1FVFfWbAiAim8Sk4eLrhFP9NbnTL7GOgVgTnURlzZVeEhNd9vBT1cY2Gp51Nn2QREam0heb4UKG9n6fxAUxPCo3avQUhjvVSfxcKn8wz7qSHCNOYCquQ=="
	//bytes, e = base64.StdEncoding.DecodeString(signString)
	//e = VerifyPKCS1v15([]byte(source), bytes, []byte(publicKey), crypto.SHA256)
	//if e != nil {
	//	fmt.Println("verify source error: ", e.Error())
	//}
}
