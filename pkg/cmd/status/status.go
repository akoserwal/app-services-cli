package status

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redhat-developer/app-services-cli/pkg/cmd/flag"
	flagutil "github.com/redhat-developer/app-services-cli/pkg/cmdutil/flags"
	"github.com/redhat-developer/app-services-cli/pkg/connection"
	"github.com/redhat-developer/app-services-cli/pkg/localize"

	"github.com/redhat-developer/app-services-cli/pkg/cmdutil/flags"

	"github.com/redhat-developer/app-services-cli/internal/config"
	"github.com/redhat-developer/app-services-cli/pkg/dump"
	"github.com/redhat-developer/app-services-cli/pkg/iostreams"
	"github.com/redhat-developer/app-services-cli/pkg/logging"
	pkgStatus "github.com/redhat-developer/app-services-cli/pkg/status"

	"github.com/redhat-developer/app-services-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	kafkaSvcName = "kafka"
)

var validServices = []string{kafkaSvcName}

type Options struct {
	IO         *iostreams.IOStreams
	Config     config.IConfig
	Logger     func() (logging.Logger, error)
	Connection factory.ConnectionFunc
	localizer  localize.Localizer

	outputFormat string
	services     []string
}

func NewStatusCommand(f *factory.Factory) *cobra.Command {
	opts := &Options{
		IO:         f.IOStreams,
		Config:     f.Config,
		Connection: f.Connection,
		Logger:     f.Logger,
		services:   validServices,
		localizer:  f.Localizer,
	}

	cmd := &cobra.Command{
		Use:       opts.localizer.MustLocalize("status.cmd.use"),
		Short:     opts.localizer.MustLocalize("status.cmd.shortDescription"),
		Long:      opts.localizer.MustLocalize("status.cmd.longDescription"),
		Example:   opts.localizer.MustLocalize("status.cmd.example"),
		ValidArgs: validServices,
		Args:      cobra.RangeArgs(0, len(validServices)),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				for _, s := range args {
					if !flags.IsValidInput(s, validServices...) {
						return errors.New(opts.localizer.MustLocalize("status.error.args.error.unknownServiceError", localize.NewEntry("ServiceName", s)))
					}
				}

				opts.services = args
			}

			validOutputFormats := flagutil.ValidOutputFormats
			if opts.outputFormat != "" && !flagutil.IsValidInput(opts.outputFormat, validOutputFormats...) {
				return flag.InvalidValueError("output", opts.outputFormat, validOutputFormats...)
			}

			return runStatus(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "", opts.localizer.MustLocalize("status.flag.output.description"))

	flagutil.EnableOutputFlagCompletion(cmd)

	return cmd
}

func runStatus(opts *Options) error {
	connection, err := opts.Connection(connection.DefaultConfigSkipMasAuth)
	if err != nil {
		return err
	}

	pkgOpts := &pkgStatus.Options{
		Config:     opts.Config,
		Connection: connection,
		Logger:     opts.Logger,
		Services:   opts.services,
	}

	logger, err := opts.Logger()
	if err != nil {
		return err
	}

	if len(opts.services) > 0 {
		logger.Debug(opts.localizer.MustLocalize("status.log.debug.requestingStatusOfServices"), opts.services)
	}

	status, ok, err := pkgStatus.Get(context.Background(), pkgOpts)
	if err != nil {
		return err
	}

	if !ok {
		logger.Info("")
		logger.Info(opts.localizer.MustLocalize("status.log.info.noStatusesAreUsed"))
		return nil
	}

	stdout := opts.IO.Out
	switch opts.outputFormat {
	case "json":
		data, _ := json.Marshal(status)
		_ = dump.JSON(stdout, data)
		return nil
	case "yaml", "yml":
		data, _ := yaml.Marshal(status)
		_ = dump.YAML(stdout, data)
		return nil
	}

	pkgStatus.Print(stdout, status)

	return nil
}
