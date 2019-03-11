package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	wemo "github.com/danward79/go.wemo"
	"github.com/gen2brain/beeep"
	gops "github.com/mitchellh/go-ps"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wemowatch",
	Short: "Wemowatch watches a process and holds a Wemo device in a specific state when the process is found.",
	Run:   wemoWatch,
}

func init() {
	rootCmd.Flags().StringP("processes", "p", "", "csv of Process names")
	rootCmd.MarkFlagRequired("processes")
	rootCmd.Flags().StringP("interface", "i", "en0", "network interface")
	rootCmd.Flags().StringP("name", "n", "", "wemo device name")
	rootCmd.MarkFlagRequired("name")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func wemoWatch(cmd *cobra.Command, args []string) {
	name := cmd.Flag("name").Value.String()
	networkInterface := cmd.Flag("interface").Value.String()
	processes := strings.Split(cmd.Flag("processes").Value.String(), ",")

	api, _ := wemo.NewByInterface(networkInterface)

	device := wemo.Device{}
	for device.Host == "" {
		devices, _ := api.DiscoverAll(3 * time.Second)
		for _, d := range devices {
			deviceInfo, _ := d.FetchDeviceInfo(context.TODO())
			if deviceInfo == nil {
				continue
			}
			log.Printf("Found device %s at %s full deviceInfo is %s\n", deviceInfo.FriendlyName, device.Host, deviceInfo)
			if strings.Contains(deviceInfo.FriendlyName, name) {
				device = *d
			}
		}
		if device.Host == "" {
			time.Sleep(60 * time.Second)
		}
	}
	deviceInfo, err := device.FetchDeviceInfo(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using device %s at %s\n", deviceInfo.FriendlyName, device.Host)

	// Loop and watch for the process
	var i int
	var lastState int
	for true {
		if i%1000 == 0 {
			lastState = device.GetBinaryState()
		}
		desiredState := 0
		systemProcesses, err := gops.Processes()
		if err != nil {
			log.Fatal(err)
		}
		for _, sp := range systemProcesses {
			spName := sp.Executable()
			for _, p := range processes {
				if strings.EqualFold(strings.ToLower(spName), strings.ToLower(p)) {
					desiredState = 1
					break
				}
			}
		}
		if lastState != desiredState {
			desiredBool := desiredState == 1
			message := fmt.Sprintf("Switching light to %t\n", desiredBool)
			log.Println(message)
			beeep.Alert("WemoWatch"+" - "+name, message, "path/to/icon.png")
			err := device.SetState(desiredBool) // Requires a bool instead of an int
			if err != nil {
				log.Fatal(err)
			}
			lastState = device.GetBinaryState()
			log.Printf("Setting lastState to %d\n", lastState)
		}
		time.Sleep(5 * time.Second)
	}

	api.Off(name, 3*time.Second)

}
