package main

import (
	"fmt"
	"strings"

	"github.com/DuckBroApprentice/chatroom/global"
)

func filterSensitive(content string) string {
	for _, word := range global.SensitiveWords {
		content = strings.ReplaceAll(content, word, "**")
	}

	return content
}

func main() {

	fmt.Println(filterSensitive("dog"))

}
