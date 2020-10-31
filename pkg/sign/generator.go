package sign

import "github.com/blademainer/commons/pkg/util"

// GenerateMd5Key generate md5 key
func GenerateMd5Key(length int) string {
	return util.RandString(length)
}

// GenerateMd5KeyWith32Word generate md5 key of random string, length is 32
func GenerateMd5KeyWith32Word() string {
	return GenerateMd5Key(32)
}
