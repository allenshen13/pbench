package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"pbench/cmd/forward"
	"pbench/utils"
	"time"
)

var forwardCmd = &cobra.Command{
	Use:                   `forward [flags]`,
	DisableFlagsInUseLine: true,
	Run:                   forward.Run,
	Args: func(cmd *cobra.Command, args []string) error {
		utils.ExpandHomeDirectory(&forward.OutputPath)
		if len(forward.PrestoFlagsArray.ServerUrl) < 2 {
			return fmt.Errorf("information for at least two clusters is required to do workload forwarding")
		}
		var sourceUrl *url.URL
		for i, serverUrl := range forward.PrestoFlagsArray.ServerUrl {
			parsedUrl, err := url.Parse(serverUrl)
			if err != nil {
				return fmt.Errorf("failed to parse server URL at position %d: %w", i, err)
			}
			if i == 0 {
				sourceUrl = parsedUrl
			} else if parsedUrl.Host == sourceUrl.Host {
				return fmt.Errorf("the forward target server host at position %d is identical to the source server host %s", i, sourceUrl.Host)
			}
		}
		for _, isTrino := range forward.PrestoFlagsArray.IsTrino {
			if isTrino {
				return fmt.Errorf("forward command does not support Trino yet")
			}
		}
		return nil
	},
	Short: "Watch incoming query workloads from the first Presto cluster (cluster 0) and forward them to the rest clusters.",
}

func init() {
	RootCmd.AddCommand(forwardCmd)
	forward.PrestoFlagsArray.Install(forwardCmd)
	_ = forwardCmd.Flags().MarkHidden("trino")
	wd, _ := os.Getwd()
	forwardCmd.Flags().BoolVarP(&forward.DryRun, "dry-run", "", false, "Turning on dry run will only show the queries but not sending them to the target server.")
	forwardCmd.Flags().StringVarP(&forward.OutputPath, "output-path", "o", wd, "Output directory path")
	forwardCmd.Flags().StringVarP(&forward.RunName, "name", "n", fmt.Sprintf("forward_%s", time.Now().Format(utils.DirectoryNameTimeFormat)), `Assign a name to this run. (default: "forward_<current time>")`)
	forwardCmd.Flags().DurationVarP(&forward.PollInterval, "poll-interval", "i", time.Second*5, "Interval between polls to the source cluster")
	forwardCmd.Flags().StringArrayVarP(&forward.ExcludePatternStrings, "exclude", "x", []string{}, `Regular expressions to filter queries to forward`)
	forwardCmd.Flags().StringArrayVarP(&forward.ReplacePatternStrings, "replace", "r", []string{}, `Pairs of regular expressions to match pattern in the query and the replacement expression. Use $1, $2, ... to reference capture groups. This will be applied after filters.`)
	forwardCmd.Flags().StringSliceVarP(&forward.SchemaMappingStrings, "schema-mapping", "m", []string{}, "Pairs of schema names to establish schema mapping relationships for schema replacement when forwarding queries. You can specify something like -m schema1,schema2")
}
