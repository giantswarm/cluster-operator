package key

import "strings"

func IsAWS(provider string) bool {
	return provider == "aws"
}

func IsAWSChina(region string) bool {
	return strings.HasPrefix(region, "cn-")
}
