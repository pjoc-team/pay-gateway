package sign

import "github.com/blademainer/commons/pkg/util"

func GenerateMd5Key(length int) string {
	return util.RandString(length)
}

func GenerateMd5KeyWith32Word() string {
	return GenerateMd5Key(32)
}
