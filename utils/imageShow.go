package utils

import (
	"fmt"
	"strings"
)

func ImageShow(index int, modelName, url string) string {
	index++
	url = strings.TrimSpace(url)
	return fmt.Sprintf("![%s](%s)", modelName, url)
}
