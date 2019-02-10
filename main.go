package main

import (
	"fmt"
	"github.com/danward79/go.wemo"
	"time"
)

func main() {
  api, _ := wemo.NewByInterface("eth0")
  devices, _ := api.DiscoverAll(30*time.Second)
  for _, device := range devices {
    fmt.Printf("Found %+v\n", device)
  }
}
