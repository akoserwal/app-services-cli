package list

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/bf2fc6cc711aee1a0c2a/cli/internal/config"
	"github.com/bf2fc6cc711aee1a0c2a/cli/internal/localizer"
	strimziadminclient "github.com/bf2fc6cc711aee1a0c2a/cli/pkg/api/strimzi-admin/client"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/cmd/factory"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/cmd/flag"
	flagutil "github.com/bf2fc6cc711aee1a0c2a/cli/pkg/cmdutil/flags"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/connection"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/dump"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/iostreams"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/logging"
	"github.com/spf13/cobra"

	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/kafka/consumergroup"
)

type Options struct {
	Config     config.IConfig
	Connection factory.ConnectionFunc
	Logger     func() (logging.Logger, error)
	IO         *iostreams.IOStreams

	output  string
	kafkaID string
	limit   int32
}

type consumerGroupRow struct {
	ConsumerGroupID   string `json:"groupId,omitempty" header:"Consumer group ID"`
	ActiveMembers     int    `json:"active_members,omitempty" header:"Active members"`
	PartitionsWithLag int    `json:"lag,omitempty" header:"Partitions with lag"`
}

// NewListConsumerGroupCommand creates a new command to list consumer groups
func NewListConsumerGroupCommand(f *factory.Factory) *cobra.Command {
	opts := &Options{
		Config:     f.Config,
		Connection: f.Connection,
		Logger:     f.Logger,
		IO:         f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:     localizer.MustLocalizeFromID("kafka.consumerGroup.list.cmd.use"),
		Short:   localizer.MustLocalizeFromID("kafka.consumerGroup.list.cmd.shortDescription"),
		Long:    localizer.MustLocalizeFromID("kafka.consumerGroup.list.cmd.longDescription"),
		Example: localizer.MustLocalizeFromID("kafka.consumerGroup.list.cmd.example"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.output != "" && !flagutil.IsValidInput(opts.output, flagutil.ValidOutputFormats...) {
				return flag.InvalidValueError("output", opts.output, flagutil.ValidOutputFormats...)
			}

			cfg, err := opts.Config.Load()
			if err != nil {
				return err
			}

			if !cfg.HasKafka() {
				return fmt.Errorf(localizer.MustLocalizeFromID("kafka.consumerGroup.common.error.noKafkaSelected"))
			}

			opts.kafkaID = cfg.Services.Kafka.ClusterID

			return runList(opts)
		},
	}

	cmd.Flags().Int32VarP(&opts.limit, "limit", "", 1000, localizer.MustLocalizeFromID("kafka.consumerGroup.list.flag.limit"))

	cmd.Flags().StringVarP(&opts.output, "output", "o", "", localizer.MustLocalize(&localizer.Config{
		MessageID:   "kafka.consumerGroup.common.flag.output.description",
		PluralCount: 2,
	}))

	return cmd
}

func runList(opts *Options) (err error) {

	conn, err := opts.Connection(connection.DefaultConfigRequireMasAuth)
	if err != nil {
		return err
	}

	logger, err := opts.Logger()
	if err != nil {
		return err
	}

	ctx := context.Background()

	api, kafkaInstance, err := conn.API().TopicAdmin(opts.kafkaID)
	if err != nil {
		return err
	}

	consumerGroupData, httpRes, consumerGroupErr := api.GetConsumerGroupList(ctx).Limit(opts.limit).Execute()

	if consumerGroupErr.Error() != "" {
		if httpRes == nil {
			return consumerGroupErr
		}

		switch httpRes.StatusCode {
		case 401:
			return errors.New(localizer.MustLocalize(&localizer.Config{
				MessageID:   "kafka.consumerGroup.common.error.unauthorized",
				PluralCount: 2,
				TemplateData: map[string]interface{}{
					"Operation": "list",
				},
			}))
		case 403:
			return errors.New(localizer.MustLocalize(&localizer.Config{
				MessageID:   "kafka.consumerGroup.common.error.forbidden",
				PluralCount: 2,
				TemplateData: map[string]interface{}{
					"Operation": "list",
				},
			}))
		case 500:
			return fmt.Errorf("%v: %w", localizer.MustLocalizeFromID("kafka.consumerGroup.common.error.internalServerError"), consumerGroupErr)
		case 503:
			return fmt.Errorf("%v: %w", localizer.MustLocalize(&localizer.Config{
				MessageID: "kafka.consumerGroup.common.error.unableToConnectToKafka",
				TemplateData: map[string]interface{}{
					"Name": kafkaInstance.GetName(),
				},
			}), consumerGroupErr)
		default:
			return consumerGroupErr
		}
	}

	if consumerGroupData.GetCount() == 0 {
		logger.Info(localizer.MustLocalize(&localizer.Config{
			MessageID: "kafka.consumerGroup.list.log.info.noConsumerGroups",
			TemplateData: map[string]interface{}{
				"InstanceName": kafkaInstance.GetName(),
			},
		}))

		return nil
	}

	stdout := opts.IO.Out
	switch opts.output {
	case "json":
		data, _ := json.Marshal(consumerGroupData)
		_ = dump.JSON(stdout, data)
	case "yaml", "yml":
		data, _ := yaml.Marshal(consumerGroupData)
		_ = dump.YAML(stdout, data)
	default:
		topics := consumerGroupData.GetItems()
		rows := mapConsumerGroupResultsToTableFormat(topics)
		dump.Table(stdout, rows)

		return nil
	}

	return nil

}

func mapConsumerGroupResultsToTableFormat(consumerGroups []strimziadminclient.ConsumerGroup) []consumerGroupRow {
	var rows []consumerGroupRow = []consumerGroupRow{}

	for _, t := range consumerGroups {
		row := consumerGroupRow{
			ConsumerGroupID:   t.GetId(),
			ActiveMembers:     len(t.GetConsumers()),
			PartitionsWithLag: consumergroup.GetPartitionsWithLag(t.GetConsumers()),
		}
		rows = append(rows, row)
	}

	return rows
}
