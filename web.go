package main // import "github.com/Dictor/jinwonbot"

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"strings"
)

func getRawHtml(path string) (string, error) {
	client := resty.New()
	resp, err := client.R().Get(path)
	return fmt.Sprint(resp), err
}

func isDoorOpen(html_string string) bool {
	if strings.Contains(html_string, "열림") {
		return true
	} else {
		return false
	}
}
