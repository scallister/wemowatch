package cmd

import (
	wemo "github.com/danward79/go.wemo"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "",
	RunE:  List,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func List(cmd *cobra.Command, args []string) (err error) {
	networkInterface := cmd.Flag("interface").Value.String()
	if networkInterface == "" {
		var bestInterfaceName string
		bestInterfaceName, err = FindBestInterfaceName()
		if err != nil {
			return
		}
		networkInterface = bestInterfaceName
	}

	wClient, err := wemo.NewByInterface(networkInterface)
	if err != nil {
		return
	}

	devices, err := wClient.DiscoverAll(TIMEOUT)
	for _, device := range devices {
		err = printDevice(device)
		if err != nil {
			return
		}
	}

	return
}
