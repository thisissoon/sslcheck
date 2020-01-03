package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"go.soon.build/sslcheck/internal/slack"

	"go.soon.build/sslcheck/internal/config"
	"go.soon.build/sslcheck/internal/ssl"
	"go.soon.build/sslcheck/internal/version"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	configkit "go.soon.build/kit/config"
)

// Default logger
var log zerolog.Logger

// Global app configuration
var cfg config.Config

// Application entry point
func main() {
	cmd := sslcheckCmd()
	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("exiting from fatal error")
		os.Exit(1)
	}
}

// New constructs a new CLI interface for execution
func sslcheckCmd() *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:           "sslcheck",
		Short:         "Check SSL certificate status for provided hosts",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			var err error
			cfg, err = config.New(
				configkit.WithFile(configPath),
				configkit.BindFlag("log.console", cmd.Flag("console")),
				configkit.BindFlag("log.verbose", cmd.Flag("verbose")),
				configkit.BindFlag("ssl.connectTimeout", cmd.Flag("timeout")),
				configkit.BindFlag("ssl.warnValidity", cmd.Flag("warning")),
				configkit.BindFlag("ssl.criticalValidity", cmd.Flag("critical")),
				configkit.BindFlag("slack.enabled", cmd.Flag("slack")),
			)
			if err != nil {
				return err
			}
			// Setup default logger
			log = initLogger(cfg.Log)
			return nil
		},
		RunE: sslcheckRun,
	}
	// Global flags
	pflags := cmd.PersistentFlags()
	pflags.StringVarP(&configPath, "config", "c", "", "path to configuration file (default is $HOME/.config/sslcheck.toml)")
	pflags.Bool("console", false, "use console log writer")
	pflags.BoolP("verbose", "v", false, "verbose logging")

	pflags.StringArrayVar(&hosts, "host", []string{}, "the domain names of the hosts to check")
	pflags.Duration("timeout", 30*time.Second, "connection timeout")
	pflags.Bool("slack", false, "send result to slack webhook")
	pflags.Int("warning", 30, "warning validity in days")
	pflags.Int("critical", 14, "critical validity in days")
	// Add sub commands
	cmd.AddCommand(versionCmd())
	return cmd
}

var hosts []string

// sslcheckRun is executed when the CLI executes
// the sslcheck command
func sslcheckRun(cmd *cobra.Command, _ []string) error {
	if hosts == nil {
		return errors.New("--host is required")
	}
	if cfg.SSL.WarnValidity < cfg.SSL.CriticalValidity {
		return errors.New("--critical is higher than --warning, i guess thats a bad idea")
	}

	c := ssl.CheckerConfig{
		ConnectTimeout:   cfg.SSL.ConnectTimeout,
		WarnValidity:     time.Hour * 24 * time.Duration(cfg.SSL.WarnValidity),
		CriticalValidity: time.Hour * 24 * time.Duration(cfg.SSL.CriticalValidity),
	}
	var attachments = []slack.Attachment{}
	for _, host := range hosts {
		certs, err := ssl.Check(log, host, c)
		if err != nil {
			return err
		}
		for _, status := range certs {
			logger := log.Info()
			switch status.Status {
			case ssl.StatusWarning, ssl.StatusCritical:
				logger = log.Warn()
			}
			msg := fmt.Sprintf("%s - expires in %s", status.CommonName, formatDuration(status.TimeRemaining))
			logger.
				Str("host", status.Host).
				Str("commonName", status.CommonName).
				Time("expiry", status.Expires).
				Dur("remainingTime", status.TimeRemaining).
				Int("status", int(status.Status)).
				Str("issuer", status.Issuer).
				Msg(msg)
			attachments = append(attachments, slack.NewAttachment(msg, int(status.Status)))
		}
	}
	// send a message to slack
	if cfg.Slack.Enabled {
		msg := slack.NewMsg(attachments)
		err := slack.SendMsg(msg, cfg.Slack.HookUrl)
		if err != nil {
			return err
		}
	}
	return nil
}

// initLogger constructs a default logger from config
func initLogger(c config.Log) zerolog.Logger {
	// Set logger level field to severity for stack driver support
	zerolog.LevelFieldName = "severity"
	var w io.Writer = os.Stdout
	if c.Console {
		w = zerolog.ConsoleWriter{
			Out: os.Stdout,
		}
	}
	// Parse level from config
	lvl, err := zerolog.ParseLevel(c.Level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	// Override level with verbose
	if c.Verbose {
		lvl = zerolog.DebugLevel
	}
	return zerolog.New(w).Level(lvl).With().Fields(map[string]interface{}{
		"version": version.Version,
		"app":     config.APP_NAME,
	}).Timestamp().Logger()
}

func formatDuration(in time.Duration) string {
	var daysPart, hoursPart, minutesPart, secondsPart string

	days := math.Floor(in.Hours() / 24)
	hoursRemaining := math.Mod(in.Hours(), 24)
	if days > 0 {
		daysPart = fmt.Sprintf("%.fd", days)
	} else {
		daysPart = ""
	}

	hours, hoursRemaining := math.Modf(hoursRemaining)
	minutesRemaining := hoursRemaining * 60
	if hours > 0 {
		hoursPart = fmt.Sprintf("%.fh", hours)
	} else {
		hoursPart = ""
	}

	if minutesRemaining > 0 {
		minutesPart = fmt.Sprintf("%.fm", minutesRemaining)
	}

	_, minutesRemaining = math.Modf(minutesRemaining)
	secondsRemaining := minutesRemaining * 60
	if secondsRemaining > 0 {
		secondsPart = fmt.Sprintf("%.fs", secondsRemaining)
	}

	return fmt.Sprintf("%s %s %s %s", daysPart, hoursPart, minutesPart, secondsPart)
}
