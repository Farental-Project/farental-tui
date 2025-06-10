package helper

import (
	"strings"
)

func RemoveEmptyLines(text string, linelimit int) string {
	var ret []string
	var arr []string

	arr = strings.Split(text, "\n")

	ret = make([]string, 0)

	for i := 0; i < len(arr); i++ {
		if len(arr[i]) > 0 {
			ret = append(ret, arr[i])
		}

		if linelimit > 0 && len(ret) >= linelimit {
			break
		}
	}

	return strings.Join(ret, "\n")
}
