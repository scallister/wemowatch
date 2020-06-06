package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "wemowatch",
	Short:             "Wemowatch watches a process and holds a Wemo device in a specific state when the process is found.",
	RunE:              watch,
	PersistentPreRunE: globalSetup,
}

func init() {
	rootCmd.Flags().StringP("processes", "p", "", "csv of Process names")
	rootCmd.MarkFlagRequired("processes")
	rootCmd.PersistentFlags().StringVarP(&interfaceName, "interface", "i", "", "network interface")
	rootCmd.Flags().StringP("name", "n", "", "wemo device name")
	rootCmd.MarkFlagRequired("name")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "")
}

// TIMEOUT - Global timeout value
var TIMEOUT = 3 * time.Second

var interfaceName string

// SharedDesiredState - Shared variable to control desired state
var SharedDesiredState int

// ActualState - Indicates the actual state
var ActualState int

// Execute - main method for cobra
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func globalSetup(cmd *cobra.Command, args []string) (err error) {
	if interfaceName == "" {
		interfaceName, err = FindBestInterfaceName()
		if err != nil {
			return
		}
	}

	if cmd.Flag("verbose").Value.String() != "true" {
		log.SetOutput(ioutil.Discard)
	}

	return
}

func watch(cmd *cobra.Command, args []string) (err error) {

	// Exit if WemoWatch is already running
	alreadyRunning, err := wemoWatchAlreadyRunning()
	if err != nil {
		return
	}
	if alreadyRunning {
		fmt.Printf("Wemo Watch is already running\n")
		os.Exit(0)
	}

	name := cmd.Flag("name").Value.String()

	processes := strings.Split(cmd.Flag("processes").Value.String(), ",")

	log.Printf("Finding device \"%s\"\n", name)
	device, err := getDeviceByName(name)
	if err != nil {
		return
	}
	deviceString, err := deviceToString(device)
	if err != nil {
		return
	}
	log.Printf("Found device: %s\n", deviceString)

	go pollActualState(device)
	go pollDesiredVsActualState(device)
	go pollIfProcessRunning(processes, device)

	log.Printf("Ready")

	// Infinite loop so the go routines can continue
	for {
		time.Sleep(1 * time.Hour)
	}

	return
}
