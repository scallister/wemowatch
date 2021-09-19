package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/soellman/pidfile"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "wemowatch",
	Short:             "Wemowatch watches a process and holds a Wemo device in a specific state when the process is found.",
	RunE:              watchDaemon,
	PersistentPreRunE: globalSetup,
}

func init() {
	rootCmd.Flags().StringP("processes", "p", "", "csv of Process names")
	err := rootCmd.MarkFlagRequired("processes")
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	rootCmd.PersistentFlags().StringVarP(&interfaceName, "interface", "i", "", "network interface")
	rootCmd.Flags().StringP("name", "n", "", "wemo device name")
	err = rootCmd.MarkFlagRequired("name")
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Set zerolog log level (error/warn/info/debug/trace")
	rootCmd.PersistentFlags().String("log-file", "", "Which file should logs be written to?")
}

// TIMEOUT - Global timeout value
var TIMEOUT = 3 * time.Second

var interfaceName string

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
	logFilePath, err := cmd.Flags().GetString("log-file")
	if err != nil {
		return err
	}
	if logFilePath != "" {
		f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			log.Error().Err(err).Msg("Failed to open log file")
			return err
		}
		fileLogger := zerolog.ConsoleWriter{Out: f, NoColor: true}
		writers = append(writers, fileLogger)
	}

	multiLevelWriter := zerolog.MultiLevelWriter(writers...)
	logger := zerolog.New(multiLevelWriter).With().Timestamp().Int("pid", os.Getpid()).Logger()
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

func watchDaemon(cmd *cobra.Command, args []string) (err error) {
	// Check for existing pid
	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		log.Error().Err(err).Msg("")

		return err
	}
	pidPath := filepath.Join(homeDirPath, ".wemowatch.pid")

	err = pidfile.WriteControl(pidPath, os.Getpid(), true)
	if errors.Is(err, pidfile.ErrProcessRunning) {
		fmt.Printf("Wemowatch is already running: %s", err.Error())
		os.Exit(0)
	}
	if err != nil {
		log.Error().Err(err).Msg("")

		return err
	}
	defer func() {
		err = pidfile.Remove(pidPath)
		if err != nil {
			log.Error().Err(err).Msg("")

			return
		}
	}()

	// Get flags
	name := cmd.Flag("name").Value.String()
	processes := strings.Split(cmd.Flag("processes").Value.String(), ",")

	// Watch processes
	for {
		err = watch(name, processes)
		if err != nil {
			log.Error().Err(err).Msg("Restarting routines in 10 minutes because of failure")
		}
		time.Sleep(10 * time.Minute)
	}
}

func watch(name string, processes []string) (err error) {
	log.Info().Msgf("Finding device \"%s\"", name)
	device, err := getDeviceByName(name)
	if err != nil {
		return err
	}
	deviceString, err := deviceToString(device)
	if err != nil {
		return err
	}
	log.Info().Msgf("Found device: %s", deviceString)

	ctx, cancel := context.WithCancel(context.Background())

	processRunning := make(chan bool)
	getActualState := make(chan bool)
	go pollActualState(ctx, getActualState)
	go pollIfProcessRunning(ctx, cancel, processRunning, processes)

	log.Info().Msg("Ready")
	var wemoIsOn bool
	for {
		select {
		case <-ctx.Done():
			log.Warn().Msg("Exiting watch because context is done")

			return nil
		case shouldBeOn := <-processRunning:
			if wemoIsOn != shouldBeOn {
				log.Info().Bool("shouldBeOn", shouldBeOn).Msg("Updating wemo device")
				err = device.SetState(shouldBeOn)
				if err != nil {
					log.Error().Err(err).Msg("")

					return err
				}
				wemoIsOn = device.GetBinaryState() == 1
			}
		case <-getActualState:
			wemoIsOn = device.GetBinaryState() == 1
		}

	}
}
