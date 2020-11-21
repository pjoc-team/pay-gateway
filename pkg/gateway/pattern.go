package gateway

import (
	"fmt"
	"github.com/pjoc-team/tracing/logger"
	"regexp"
	"strings"
)

var regexMap = make(map[string]*regexp.Regexp)

// ReplaceGatewayOrderID replace gateway order id
func ReplaceGatewayOrderID(urlPattern string, gatewayOrderID string) string {
	url, e := ReplacePlaceholder(urlPattern, "gateway_order_id", gatewayOrderID)
	if e != nil {
		return strings.Replace(urlPattern, "{gateway_order_id}", gatewayOrderID, -1)
	}
	return url
}

// ReplacePlaceholder replace place holder by parameter
func ReplacePlaceholder(urlPattern string, placeHolderName string, parameter string) (string, error) {
	compile, e := GetPlaceholderRegex(placeHolderName)
	if e != nil {
		logger.Log().Errorf("Regex error: %v", e.Error())
		return "", e
	}
	result := compile.ReplaceAll([]byte(urlPattern), []byte(parameter))
	return string(result), nil
}

// GetPlaceholderRegex regex
func GetPlaceholderRegex(placeHolderName string) (*regexp.Regexp, error) {
	if regex, found := regexMap[placeHolderName]; found {
		return regex, nil
	}
	pattern := fmt.Sprintf("\\{\\s*%s\\s*\\}", placeHolderName)
	compile, e := regexp.Compile(pattern)
	if e != nil {
		logger.Log().Errorf("Regex error: %v", e.Error())
		return nil, e
	}
	regexMap[placeHolderName] = compile
	return compile, e
}
