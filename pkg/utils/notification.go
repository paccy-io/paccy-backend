package utils

import (
	"fmt"
	"net/url"
)

func GetEmailVerificationLink(token string) string {
	return fmt.Sprintf("https://paccy.io/verify-email?token=%s", url.QueryEscape(token))
}


