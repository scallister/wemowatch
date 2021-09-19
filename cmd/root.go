package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Set zerolog log level (error/warn/info/debug/trace")
	rootCmd.PersistentFlags().String("log-file", "", "Which file should logs be written to?")
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
		log.Error().Err(err).Msg("Exiting because of error")
		os.Exit(1)
	}
}

func globalSetup(cmd *cobra.Command, args []string) (err error) {
	var writers []io.Writer

	// Add console logging
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
	writers = append(writers, consoleWriter)

	// Add file logging if specified
	logFile, err := cmd.Flags().GetString("log-file")
	if err != nil {
		return err
	}
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModeAppend)
		if err != nil {
			return err
		}
		fileLogger := zerolog.ConsoleWriter{Out: f, NoColor: true}
		writers = append(writers, fileLogger)
	}

	multiLevelWriter := zerolog.MultiLevelWriter(writers...)
	logger := zerolog.New(multiLevelWriter).With().Timestamp().Logger()
	log.Logger = logger

	logLevelStr, err := cmd.Flags().GetString("log-level")
	if err != nil {
		return
	}
	logLevel, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		return
	}
	zerolog.SetGlobalLevel(logLevel)

	if interfaceName == "" {
		interfaceName, err = FindBestInterfaceName()
		if err != nil {
			return
		}
	}

	return
}

func watch(cmd *cobra.Command, args []string) (err error) {
	err = watchDaemon(cmd, args)
	if err != nil {
		log.Error().Err(err).Msg("Exiting because of this error")
	}

	return err
}

func watchDaemon(cmd *cobra.Command, args []string) (err error) {
	isAlreadyRunning, err := alreadyRunning()
	if err != nil {
		return
	}
	// Exit if WemoWatch is already running
	if isAlreadyRunning {
		msg := fmt.Sprintf("wemowatch is already running")
		err = errors.New(msg)
		log.Error().Err(err)
		return err
	}

	name := cmd.Flag("name").Value.String()

	processes := strings.Split(cmd.Flag("processes").Value.String(), ",")

	log.Info().Msgf("Finding device \"%s\"", name)
	device, err := getDeviceByName(name)
	if err != nil {
		return
	}
	deviceString, err := deviceToString(device)
	if err != nil {
		return
	}
	log.Info().Msgf("Found device: %s", deviceString)

	go pollActualState(device)
	go pollDesiredVsActualState(device)
	go pollIfProcessRunning(processes, device)

	log.Info().Msg("Ready")

	// Infinite loop so the go routines can continue
	for {
		time.Sleep(1 * time.Hour)
	}

	return
}
