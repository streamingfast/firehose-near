package tools

import (
	"strings"

	"github.com/spf13/cobra"
	pbtransform "github.com/streamingfast/sf-near/pb/sf/near/transform/v1"
	sftools "github.com/streamingfast/sf-tools"
	"google.golang.org/protobuf/types/known/anypb"
)

func init() {
	firehoseClientCmd := sftools.GetFirehoseClientCmd(zlog, tracer, transformsSetter)
	firehoseClientCmd.Flags().String("receipt-account-filters", "", "comma-separated accounts to use as filter/index. If it contains a colon (:), it will be interpreted as <prefix>:<suffix> (each of which can be empty, ex: 'hello:' or ':world')")
	Cmd.AddCommand(firehoseClientCmd)
}

var transformsSetter = func(cmd *cobra.Command) (transforms []*anypb.Any, err error) {
	filters, err := parseFilters(mustGetString(cmd, "receipt-account-filters"))
	if err != nil {
		return nil, err
	}

	if filters != nil {
		t, err := anypb.New(filters)
		if err != nil {
			return nil, err
		}
		transforms = append(transforms, t)
	}
	return
}

func parseFilters(in string) (*pbtransform.BasicReceiptFilter, error) {
	if in == "" {
		return nil, nil
	}
	var pairs []*pbtransform.PrefixSuffixPair
	var accounts []string

	for _, unit := range strings.Split(in, ",") {
		if parts := strings.Split(unit, ":"); len(parts) == 2 {
			pairs = append(pairs, &pbtransform.PrefixSuffixPair{
				Prefix: parts[0],
				Suffix: parts[1],
			})
			continue
		}
		accounts = append(accounts, unit)
	}

	filters := &pbtransform.BasicReceiptFilter{
		Accounts:             accounts,
		PrefixAndSuffixPairs: pairs,
	}

	return filters, nil
}
