package main

import "github.com/bygui86/go-tor-proxy/common"

func main() {

	_, _ = common.NewHTTPClient().Get("https://api.wind.io/forecast")
}
