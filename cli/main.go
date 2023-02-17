package main

import (
	"github.com/eduaravila/simplechat/pkg/chat"
)

func main() {
	chat := chat.NewChat()
	chat.Start()
}
