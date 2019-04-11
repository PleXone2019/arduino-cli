/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package root

import (
	"io/ioutil"
	"os"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/cli/board"
	"github.com/arduino/arduino-cli/cli/compile"
	"github.com/arduino/arduino-cli/cli/core"
	"github.com/arduino/arduino-cli/cli/daemon"
	"github.com/arduino/arduino-cli/cli/upload"
	"github.com/arduino/arduino-cli/commands/config"
	"github.com/arduino/arduino-cli/commands/generatedocs"
	"github.com/arduino/arduino-cli/commands/lib"
	"github.com/arduino/arduino-cli/commands/sketch"
	"github.com/arduino/arduino-cli/commands/version"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/configs"
	paths "github.com/arduino/go-paths-helper"
	colorable "github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// Init prepares the cobra root command.
func Init() *cobra.Command {
	command := &cobra.Command{
		Use:              "arduino-cli",
		Short:            "Arduino CLI.",
		Long:             "Arduino Command Line Interface (arduino-cli).",
		Example:          "  " + cli.AppName + " <command> [flags...]",
		PersistentPreRun: preRun,
	}
	command.PersistentFlags().BoolVar(&cli.GlobalFlags.Debug, "debug", false, "Enables debug output (super verbose, used to debug the CLI).")
	command.PersistentFlags().StringVar(&outputFormat, "format", "text", "The output format, can be [text|json].")
	command.PersistentFlags().StringVar(&yamlConfigFile, "config-file", "", "The custom config file (if not specified ./.cli-config.yml will be used).")
	command.AddCommand(board.InitCommand())
	command.AddCommand(compile.InitCommand())
	command.AddCommand(config.InitCommand())
	command.AddCommand(core.InitCommand())
	command.AddCommand(daemon.InitCommand())
	command.AddCommand(generatedocs.InitCommand())
	command.AddCommand(lib.InitCommand())
	// command.AddCommand(login.InitCommand())
	// command.AddCommand(logout.InitCommand())
	command.AddCommand(sketch.InitCommand())
	command.AddCommand(upload.InitCommand())
	// command.AddCommand(validate.InitCommand())
	command.AddCommand(version.InitCommand())
	return command
}

var outputFormat string
var yamlConfigFile string

func preRun(cmd *cobra.Command, args []string) {
	// Reset logrus if debug flag changed.
	if !cli.GlobalFlags.Debug {
		// Discard logrus output if no debug.
		logrus.SetOutput(ioutil.Discard)
	} else {
		// Else print on stderr.

		// Workaround to get colored output on windows
		if terminal.IsTerminal(int(os.Stdout.Fd())) {
			logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
		}
		logrus.SetOutput(colorable.NewColorableStdout())
		cli.ErrLogrus.Out = colorable.NewColorableStderr()
		formatter.SetLogger(cli.ErrLogrus)
	}
	initConfigs()

	logrus.Info(cli.AppName + "-" + cli.Version)
	logrus.Info("Starting root command preparation (`arduino`)")
	switch outputFormat {
	case "text":
		formatter.SetFormatter("text")
		cli.GlobalFlags.OutputJSON = false
	case "json":
		formatter.SetFormatter("json")
		cli.GlobalFlags.OutputJSON = true
	default:
		formatter.PrintErrorMessage("Invalid output format: " + outputFormat)
		os.Exit(cli.ErrBadCall)
	}

	logrus.Info("Formatter set")
	if !formatter.IsCurrentFormat("text") {
		cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			logrus.Warn("Calling help on JSON format")
			formatter.PrintErrorMessage("Invalid Call : should show Help, but it is available only in TEXT mode.")
			os.Exit(cli.ErrBadCall)
		})
	}
}

// initConfigs initializes the configuration from the specified file.
func initConfigs() {
	if conf, err := configs.NewConfiguration(); err != nil {
		logrus.WithError(err).Error("Error creating default configuration")
		formatter.PrintError(err, "Error creating default configuration")
		os.Exit(cli.ErrGeneric)
	} else {
		cli.Config = conf
	}

	if yamlConfigFile != "" {
		cli.Config.ConfigFile = paths.New(yamlConfigFile)
	}

	logrus.Info("Initiating configuration")
	if err := cli.Config.LoadFromYAML(cli.Config.ConfigFile); err != nil {
		logrus.WithError(err).Warn("Did not manage to get config file, using default configuration")
	}
	if cli.Config.IsBundledInDesktopIDE() {
		logrus.Info("CLI is bundled into the IDE")
		err := cli.Config.LoadFromDesktopIDEPreferences()
		if err != nil {
			logrus.WithError(err).Warn("Did not manage to get config file of IDE, using default configuration")
		}
	} else {
		logrus.Info("CLI is not bundled into the IDE")
	}
	cli.Config.LoadFromEnv()
	logrus.Info("Configuration set")
}
