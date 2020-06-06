package cmd

import (
	"context"
	"fmt"
	wemo "github.com/danward79/go.wemo"
	gops "github.com/mitchellh/go-ps"
	"log"
	"net"
	"strings"
	"time"
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

// pollIfProcessRunning - Continually runs and updates `desiredStateOn`
func pollIfProcessRunning(processes []string, device *wemo.Device) (err error) {
	var desiredState int

	var processesLowerCase []string
	for _, process := range processes {
		processesLowerCase = append(processesLowerCase, strings.ToLower(process))
	}

	for {
		var systemProcesses []gops.Process
		systemProcesses, err = gops.Processes()
		if err != nil {
			return
		}

		desiredState = 0
		for _, sp := range systemProcesses {
			spName := strings.ToLower(sp.Executable())
			for _, p := range processesLowerCase {
				// One of the desired processes is running
				if strings.EqualFold(spName, strings.ToLower(p)) {
					desiredState = 1
					break
				}
			}
		}
		// Set the shared desiredStateOn variable appropriately
		if SharedDesiredState != desiredState {
			SharedDesiredState = desiredState
			log.Printf("Set desired state to: %d\n", desiredState)
			err = setState(device)
			if err != nil {
				return
			}

		}
		time.Sleep(1 * time.Minute)
	}
	return
}

// pollActualState - Confirms the actual state is what we believe it is
func pollActualState(device *wemo.Device) (err error) {
	var newState int
	for {
		newState = device.GetBinaryState()
		if newState != ActualState {
			ActualState = device.GetBinaryState()
			log.Printf("Updated actual state to: %d\n", ActualState)
			err = setState(device)
			if err != nil {
				return err
			}
		}
		time.Sleep(10 * time.Minute) // Check every 10 minutes
	}
	return
}

// changeState - Continually switch the state to the desired state if needed
func pollDesiredVsActualState(device *wemo.Device) (err error) {
	for {
		if SharedDesiredState != ActualState {
			err = setState(device)
			if err != nil {
				return
			}
		}
		time.Sleep(1 * time.Minute)
	}
	return
}

// setState - Set the state to the proper value
func setState(device *wemo.Device) (err error) {
	newState := SharedDesiredState == 1
	err = device.SetState(newState)
	if err != nil {
		log.Fatal(err) // TODO improve this
		return
	}
	log.Printf("Set to %t\n", newState)
	ActualState = device.GetBinaryState()
	return
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
