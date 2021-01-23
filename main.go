package main

import (
	"avatar/web"
	"os"
)

func main() {
	var listenAddr = ":2199"
	if addr, ok := os.LookupEnv("AVATAR_ADDR"); ok {
		listenAddr = addr
	}
	if err := web.Listen(listenAddr); err != nil {
		panic(err)
	}
}
