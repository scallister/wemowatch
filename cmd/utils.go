package cmd

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	wemo "github.com/danward79/go.wemo"
	gops "github.com/mitchellh/go-ps"
	"github.com/rs/zerolog/log"
)

func getWClient() (wClient *wemo.Wemo, err error) {
	return wemo.NewByInterface(interfaceName)
}

func deviceToString(device *wemo.Device) (result string, err error) {
	deviceInfo, err := device.FetchDeviceInfo(context.TODO())
	if err != nil {
		return
	}
	ip := strings.Split(device.Host, ":")[0]
	result = fmt.Sprintf("\"%s\" (IP: %s)",
		deviceInfo.FriendlyName,
		ip)
	return
}

func printDevice(device *wemo.Device) (err error) {
	result, err := deviceToString(device)
	if err != nil {
		return
	}
	fmt.Println(result)
	deviceInfo, err := device.FetchDeviceInfo(context.TODO())
	if err != nil {
		return
	}
	ip := strings.Split(device.Host, ":")[0]
	fmt.Printf("\"%s\" (IP: %s)\n",
		deviceInfo.FriendlyName,
		ip)
	return
}

func getDeviceByName(name string) (correctDevice *wemo.Device, err error) {
	wClient, err := getWClient()
	if err != nil {
		return
	}

	var devices []*wemo.Device
	devices, err = wClient.DiscoverAll(TIMEOUT)
	if err != nil {
		return
	}

	for _, d := range devices {
		deviceInfo, _ := d.FetchDeviceInfo(context.TODO())
		if deviceInfo == nil {
			continue
		}
		if strings.Contains(deviceInfo.FriendlyName, name) {
			correctDevice = d
			return
		}
	}

	return
}

func alreadyRunning() (alreadyRunning bool, err error) {
	systemProcesses, err := gops.Processes()
	if err != nil {
		return
	}

	count := 0
	for _, sp := range systemProcesses {
		// Check if the name matches
		spName := strings.ToLower(sp.Executable())
		if !strings.Contains(spName, "wemowatch") {
			continue
		}

		count += 1
	}
	// Needs to be greater than one so we do not count this process
	// if it is the only one running
	if count > 1 {
		alreadyRunning = true
	}
	return
}

// pollIfProcessRunning - Continually runs and updates `desiredStateOn`
func pollIfProcessRunning(ctx context.Context, cancel func(), processRunning chan bool, processes []string) {
	var processesLowerCase []string
	for _, process := range processes {
		processesLowerCase = append(processesLowerCase, strings.ToLower(process))
	}

	for {
		var err error
		var systemProcesses []gops.Process
		systemProcesses, err = gops.Processes()
		if err != nil {
			log.Error().Err(err).Msg("Exiting pollIfProcessRunning because of error and cancelling other routines")
			cancel()

			return
		}

		found := false
		for _, sp := range systemProcesses {
			spName := strings.ToLower(sp.Executable())
			subLog := log.With().Strs("desiredProcesses", processesLowerCase).Str("systemProcess", spName).Logger()
			for _, p := range processesLowerCase {
				// One of the desired processes is running
				if strings.EqualFold(spName, strings.ToLower(p)) {
					subLog.Debug().Msg("Process match found")
					found = true

					break
				}
				subLog.Trace().Msg("Processes did not match")
			}
		}
		select {
		case <-ctx.Done():
			log.Warn().Msg("Exiting pollIfProcessRunning because context is done")

			return
		case processRunning <- found:
			time.Sleep(15 * time.Second)
		}
	}
}

// pollActualState - Indicates when the actual state should be fetched
func pollActualState(ctx context.Context, getActualState chan bool) {
	for {
		select {
		case <-ctx.Done():
			log.Warn().Msg("Exiting pollActualState because context finished")
			return
		case getActualState <- true:
			time.Sleep(10 * time.Minute)
		}
	}
}

func FindBestInterfaceName() (bestInterfaceName string, err error) {
	bestInterface, err := FindBestInterface()
	if err != nil {
		return
	}
	bestInterfaceName = bestInterface.Name
	return
}

func FindBestInterface() (bestInterface net.Interface, err error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, i := range ifaces {
		var addrs []net.Addr
		addrs, err = i.Addrs()
		if err != nil {
			return
		}
		for _, addr := range addrs {
			if strings.Contains(addr.String(), "192.168") {
				bestInterface = i
				return
			}
		}
	}

	return
}
