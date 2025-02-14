package run

import (
	"fmt"
	"hitman/internal/config"
	"hitman/internal/globals"
	"hitman/internal/processor"
	"log"
	"time"

	"github.com/spf13/cobra"
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

	// Parse and store the config
	go configProcessorWorker(configPath)

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

// configProcessorWorker TODO
func configProcessorWorker(configPath string) {

	for {

		// Parse and store the config
		configContent, err := config.ReadFile(configPath)
		if err != nil {
			globals.ExecContext.Logger.Fatalf(fmt.Sprintf(ConfigNotParsedErrorMessage, err))
		}

		//
		duration, err := time.ParseDuration(configContent.Spec.Synchronization.Time)
		if err != nil {
			globals.ExecContext.Logger.Fatalf(UnableParseDurationErrorMessage, err)
		}

		//
		durationDelay, err := time.ParseDuration(configContent.Spec.Synchronization.ProcessingDelay)
		if err != nil {
			globals.ExecContext.Logger.Fatalf(UnableParseDurationErrorMessage, err)
		}

		//
		configContent.Spec.Synchronization.CarriedTime = duration
		configContent.Spec.Synchronization.CarriedProcessingDelay = durationDelay

		//
		globals.ExecContext.Config.Mutex.Lock()

		globals.ExecContext.Config.ApiVersion = configContent.ApiVersion
		globals.ExecContext.Config.Kind = configContent.Kind
		globals.ExecContext.Config.Metadata = configContent.Metadata
		globals.ExecContext.Config.Spec = configContent.Spec

		globals.ExecContext.Config.Mutex.Unlock()

		//
		time.Sleep(2 * time.Second)
	}
}
