package queue

import "strings"

// ParseBrokers 按逗号解析数组
func ParseBrokers(brokers string) []string {
	brokerArray := strings.Split(brokers, ",")
	for i, s := range brokerArray {
		brokerArray[i] = strings.TrimSpace(s)
	}
	return brokerArray
}
