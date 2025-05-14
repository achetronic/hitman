package run

import (
	"fmt"
	"hitman/api/v1alpha1"
	"log"
	"reflect"
	"time"

	//
	"github.com/spf13/cobra"

	//
	"hitman/internal/config"
	"hitman/internal/globals"
	"hitman/internal/processor"
)

const (
	descriptionShort = `Execute synchronization process`

	descriptionLong = `
	Run execute synchronization process`

	//

	//
	ConfigFlagErrorMessage          = "impossible to get flag --config: %s"
	ConfigNotParsedErrorMessage     = "impossible to parse config file: %s"
	LogLevelFlagErrorMessage        = "impossible to get flag --log-level: %s"
	DisableTraceFlagErrorMessage    = "impossible to get flag --disable-trace: %s"
	DryRunFlagErrorMessage          = "impossible to get flag --dry-run: %s"
	UnableParseDurationErrorMessage = "unable to parse duration: %s"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "run",
		DisableFlagsInUseLine: true,
		Short:                 descriptionShort,
		Long:                  descriptionLong,

		Run: RunCommand,
	}

	//
	cmd.Flags().String("log-level", "info", "Verbosity level for logs")
	cmd.Flags().Bool("disable-trace", true, "Disable showing traces in logs")
	cmd.Flags().String("config", "hitman.yaml", "Path to the YAML config file")
	cmd.Flags().Bool("dry-run", false, "Disable performing actual actions")

	return cmd
}

// RunCommand TODO
// Ref: https://pkg.go.dev/github.com/spf13/pflag#StringSlice
func RunCommand(cmd *cobra.Command, args []string) {

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatalf(ConfigFlagErrorMessage, err)
	}

	// Init the logger and store the level into the context
	logLevelFlag, err := cmd.Flags().GetString("log-level")
	if err != nil {
		log.Fatalf(LogLevelFlagErrorMessage, err)
	}
	globals.ExecContext.LogLevel = logLevelFlag

	disableTraceFlag, err := cmd.Flags().GetBool("disable-trace")
	if err != nil {
		log.Fatalf(DisableTraceFlagErrorMessage, err)
	}

	err = globals.SetLogger(logLevelFlag, disableTraceFlag)
	if err != nil {
		log.Fatal(err)
	}

	dryRunFlag, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		log.Fatalf(DryRunFlagErrorMessage, err)
	}
	globals.ExecContext.DryRun = dryRunFlag

	/////////////////////////////
	// EXECUTION FLOW RELATED
	/////////////////////////////

	globals.ExecContext.Logger.Infof("starting Hitman. Getting ready to kill some targets")

	// Parse and store the config in the background
	// Main process must wait until config is being processed, at least, once
	configReady := make(chan struct{})
	go configProcessorWorker(configPath, configReady)
	<-configReady // Wait until config is ready

	//
	processorObj, err := processor.NewProcessor()
	if err != nil {
		globals.ExecContext.Logger.Infof("error creating processor: %s", err.Error())
	}

	for {
		globals.ExecContext.Logger.Info("syncing resources")

		globals.ExecContext.Config.Mutex.RLock()
		err = processorObj.SyncResources()
		if err != nil {
			globals.ExecContext.Logger.Infof("error syncing resources: %s", err)
		}

		//
		globals.ExecContext.Logger.Infof("syncing again in %s",
			globals.ExecContext.Config.Spec.Synchronization.CarriedTime.String())
		globals.ExecContext.Config.Mutex.RUnlock()
		time.Sleep(globals.ExecContext.Config.Spec.Synchronization.CarriedTime)
	}
}

// configProcessorWorker TODO - Reads and applies configuration initially,
// then reloads periodically
func configProcessorWorker(configPath string, configReady chan<- struct{}) {
	// Initial load
	applyConfig(configPath)

	// Signal main that initial config is ready
	close(configReady)

	// Periodic reload every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		applyConfig(configPath)
	}
}

// applyConfig TODO - Reads the config file, parses durations,
// and updates globals.ExecContext.Config
func applyConfig(configPath string) {
	configContent, err := config.ReadFile(configPath)
	if err != nil {
		globals.ExecContext.Logger.Fatalf(fmt.Sprintf(ConfigNotParsedErrorMessage, err))
	}

	// Set default synchronization times if zero
	if reflect.ValueOf(configContent.Spec.Synchronization.Time).IsZero() {
		configContent.Spec.Synchronization.Time = v1alpha1.DefaultSyncTime
	}
	duration, err := time.ParseDuration(configContent.Spec.Synchronization.Time)
	if err != nil {
		globals.ExecContext.Logger.Fatalf(UnableParseDurationErrorMessage, err)
	}

	if reflect.ValueOf(configContent.Spec.Synchronization.ProcessingDelay).IsZero() {
		configContent.Spec.Synchronization.ProcessingDelay = v1alpha1.DefaultSyncProcessingDelay
	}
	durationDelay, err := time.ParseDuration(configContent.Spec.Synchronization.ProcessingDelay)
	if err != nil {
		globals.ExecContext.Logger.Fatalf(UnableParseDurationErrorMessage, err)
	}

	configContent.Spec.Synchronization.CarriedTime = duration
	configContent.Spec.Synchronization.CarriedProcessingDelay = durationDelay

	// Apply updated config under lock
	globals.ExecContext.Config.Mutex.Lock()

	globals.ExecContext.Config.ApiVersion = configContent.ApiVersion
	globals.ExecContext.Config.Kind = configContent.Kind
	globals.ExecContext.Config.Metadata = configContent.Metadata
	globals.ExecContext.Config.Spec = configContent.Spec

	globals.ExecContext.Config.Mutex.Unlock()
}
