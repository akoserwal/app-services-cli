package describe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bf2fc6cc711aee1a0c2a/cli/internal/config"
	"github.com/bf2fc6cc711aee1a0c2a/cli/internal/localizer"
	strimziadminclient "github.com/bf2fc6cc711aee1a0c2a/cli/pkg/api/strimzi-admin/client"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/cmd/factory"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/cmd/flag"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/connection"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/dump"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/iostreams"
	"github.com/bf2fc6cc711aee1a0c2a/cli/pkg/kafka/consumergroup"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Options struct {
	kafkaID      string
	outputFormat string
	id           string

	IO         *iostreams.IOStreams
	Config     config.IConfig
	Connection factory.ConnectionFunc
}

type consumerRow struct {
	MemberID      string `json:"memberId,omitempty" header:"Member ID"`
	Partition     int    `json:"partition,omitempty" header:"Partition"`
	LogEndOffset  int    `json:"logEndOffset,omitempty" header:"Log end offset"`
	CurrentOffset int    `json:"offset,omitempty" header:"Current offset"`
	OffsetLag     int    `json:"lag,omitempty" header:"Offset lag"`
}

// NewDescribeConsumerGroupCommand gets a new command for describing a consumer group.
func NewDescribeConsumerGroupCommand(f *factory.Factory) *cobra.Command {
	opts := &Options{
		Connection: f.Connection,
		Config:     f.Config,
		IO:         f.IOStreams,
	}
	cmd := &cobra.Command{
		Use:     localizer.MustLocalizeFromID("kafka.consumerGroup.describe.cmd.use"),
		Short:   localizer.MustLocalizeFromID("kafka.consumerGroup.describe.cmd.shortDescription"),
		Long:    localizer.MustLocalizeFromID("kafka.consumerGroup.describe.cmd.longDescription"),
		Example: localizer.MustLocalizeFromID("kafka.consumerGroup.describe.cmd.example"),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			opts.id = args[0]

			if opts.outputFormat != "" {
				if err = flag.ValidateOutput(opts.outputFormat); err != nil {
					return err
				}
			}

			if opts.kafkaID != "" {
				return runCmd(opts)
			}

			cfg, err := opts.Config.Load()
			if err != nil {
				return err
			}

			if !cfg.HasKafka() {
				return errors.New(localizer.MustLocalizeFromID("kafka.consumerGroup.common.error.noKafkaSelected"))
			}

			opts.kafkaID = cfg.Services.Kafka.ClusterID

			return runCmd(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.outputFormat, "output", "o", "", localizer.MustLocalize(&localizer.Config{
		MessageID: "kafka.consumerGroup.common.flag.output.description",
	}))

	return cmd
}

func runCmd(opts *Options) error {
	conn, err := opts.Connection(connection.DefaultConfigRequireMasAuth)
	if err != nil {
		return err
	}

	api, kafkaInstance, err := conn.API().TopicAdmin(opts.kafkaID)
	if err != nil {
		return err
	}

	ctx := context.Background()

	consumerGroupData, httpRes, consumerGroupErr := api.GetConsumerGroupById(ctx, opts.id).Execute()

	if consumerGroupErr.Error() != "" {
		if httpRes == nil {
			return consumerGroupErr
		}

		switch httpRes.StatusCode {
		case 401:
			return errors.New(localizer.MustLocalize(&localizer.Config{
				MessageID: "kafka.consumerGroup.common.error.unauthorized",
				TemplateData: map[string]interface{}{
					"Operation": "view",
				},
			}))
		case 403:
			return errors.New(localizer.MustLocalize(&localizer.Config{
				MessageID: "kafka.consumerGroup.common.error.forbidden",
				TemplateData: map[string]interface{}{
					"Operation": "view",
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

	stdout := opts.IO.Out
	switch opts.outputFormat {
	case "json":
		data, _ := json.Marshal(consumerGroupData)
		_ = dump.JSON(stdout, data)
	case "yaml", "yml":
		data, _ := yaml.Marshal(consumerGroupData)
		_ = dump.YAML(stdout, data)
	default:
		consumers := consumerGroupData.GetConsumers()
		rows := mapConsumerGroupDescribeToTableFormat(consumers)
		fmt.Fprintln(stdout, localizer.MustLocalize(&localizer.Config{
			MessageID: "kafka.consumerGroup.describe.output.id",
			TemplateData: map[string]interface{}{
				"ID": consumerGroupData.GetId(),
			},
		}))
		fmt.Fprintln(stdout, localizer.MustLocalize(&localizer.Config{
			MessageID: "kafka.consumerGroup.describe.output.activeMembers",
			TemplateData: map[string]interface{}{
				"ActiveMembers": len(consumerGroupData.GetConsumers()),
			},
		}))
		fmt.Fprintln(stdout, localizer.MustLocalize(&localizer.Config{
			MessageID: "kafka.consumerGroup.describe.output.partitionsWithLag",
			TemplateData: map[string]interface{}{
				"LaggingPartitions": consumergroup.GetPartitionsWithLag(consumerGroupData.GetConsumers()),
			},
		}))
		fmt.Fprintln(stdout, "")
		dump.Table(stdout, rows)
	}

	return nil
}

func mapConsumerGroupDescribeToTableFormat(consumers []strimziadminclient.Consumer) []consumerRow {

	var rows []consumerRow = []consumerRow{}

	for _, consumer := range consumers {

		row := consumerRow{
			Partition:     int(consumer.GetPartition()),
			MemberID:      consumer.GetMemberId(),
			LogEndOffset:  int(consumer.GetLogEndOffset()),
			CurrentOffset: int(consumer.GetOffset()),
			OffsetLag:     int(consumer.GetLag()),
		}
		rows = append(rows, row)
	}

	return rows
}
