package run

import (
	"fmt"
	"hitman/internal/config"
	"hitman/internal/globals"
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
	SyncTimeFlagErrorMessage        = "impossible to get flag --sync-time: %s"
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
	cmd.Flags().String("sync-time", "10m", "Waiting time between group synchronizations (in duration type)")
	cmd.Flags().String("config", "hitman.yaml", "Path to the YAML config file")

	return cmd
}

// RunCommand TODO
// Ref: https://pkg.go.dev/github.com/spf13/pflag#StringSlice
func RunCommand(cmd *cobra.Command, args []string) {

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		log.Fatalf(ConfigFlagErrorMessage, err)
	}

	// Init the logger
	logLevelFlag, err := cmd.Flags().GetString("log-level")
	if err != nil {
		log.Fatalf(LogLevelFlagErrorMessage, err)
	}

	disableTraceFlag, err := cmd.Flags().GetBool("disable-trace")
	if err != nil {
		log.Fatalf(DisableTraceFlagErrorMessage, err)
	}

	err = globals.SetLogger(logLevelFlag, disableTraceFlag)
	if err != nil {
		log.Fatal(err)
	}

	//
	syncTime, err := cmd.Flags().GetString("sync-time")
	if err != nil {
		globals.ExecContext.Logger.Fatalf(SyncTimeFlagErrorMessage, err)
	}

	/////////////////////////////
	// EXECUTION FLOW RELATED
	/////////////////////////////

	// Get and parse the config
	configContent, err := config.ReadFile(configPath)
	if err != nil {
		globals.ExecContext.Logger.Fatalf(fmt.Sprintf(ConfigNotParsedErrorMessage, err))
	}

	_ = configContent // TODO

	//
	globals.ExecContext.Logger.Infof("Starting Hitman. Ready to kill some targets")

	for {

		log.Print("Testing main flow")

		//
		duration, err := time.ParseDuration(syncTime)
		if err != nil {
			globals.ExecContext.Logger.Fatalf(UnableParseDurationErrorMessage, err)
		}
		globals.ExecContext.Logger.Infof("Syncing again in %s", duration.String())
		time.Sleep(duration)
	}
}
