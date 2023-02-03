package tools

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mr-tron/base58"
	"github.com/spf13/cobra"
	"github.com/streamingfast/cli"
)

func init() {
	Cmd.AddCommand(GenerateNodeKeyCmd)
}

var GenerateNodeKeyCmd = &cobra.Command{
	Use:   "generate-node-key [<output_file>]",
	Short: "Generate a new node key JSON file suitable to be used by NEAR node, if no argument is provided, write to './node_key.json'",
	Args:  cobra.RangeArgs(0, 1),
	RunE:  generateNodeKeyE,
	Example: string(cli.ExamplePrefixed("firenear tools", `
		# Generate NEAR node key in file named 'node_key.json' in current directory
		generate-node-key

		# Generate NEAR node key in file named 'path/node_key.json'
		generate-node-key path/my_node_key.json
	`)),
}

func generateNodeKeyE(cmd *cobra.Command, args []string) error {
	type NodeKey struct {
		AccountID  string `json:"account_id"`
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}

	outputFile := "node_key.json"
	if len(args) > 0 {
		outputFile = args[0]
	}

	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	cli.NoError(err, "Unable to generate ed25519 public/private key pair")

	out, err := json.MarshalIndent(NodeKey{
		AccountID:  "node",
		PublicKey:  fmt.Sprintf("ed25519:%s", base58.Encode([]byte(publicKey))),
		PrivateKey: fmt.Sprintf("ed25519:%s", base58.Encode([]byte(privateKey))),
	}, "", "  ")
	cli.NoError(err, "Unable to marshal node key")

	err = os.WriteFile(outputFile, out, os.ModePerm)
	cli.NoError(err, "Unable to write output file")

	return err
}
