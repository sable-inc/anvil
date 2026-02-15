package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sable-inc/anvil/internal/output"
)

func newRawAPICmd() *cobra.Command {
	var data string

	cmd := &cobra.Command{
		Use:   "api <METHOD> <path>",
		Short: "Make a raw API request",
		Long:  "Send an arbitrary HTTP request to the Sable API.\nThe auth token is injected automatically.",
		Example: `  # GET request
  anvil api GET /agents

  # POST with JSON body
  anvil api POST /agents -d '{"name":"my-agent"}'

  # DELETE
  anvil api DELETE /agents/123`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			a := AppFrom(cmd)
			client, err := a.RequireAuth()
			if err != nil {
				return err
			}

			method := strings.ToUpper(args[0])
			path := args[1]

			var body any
			if data != "" {
				if err := json.Unmarshal([]byte(data), &body); err != nil {
					return fmt.Errorf("invalid JSON body: %w", err)
				}
			}

			var resp any
			switch method {
			case "GET":
				err = client.Get(cmd.Context(), path, &resp)
			case "POST":
				err = client.Post(cmd.Context(), path, body, &resp)
			case "PUT":
				err = client.Put(cmd.Context(), path, body, &resp)
			case "PATCH":
				err = client.Patch(cmd.Context(), path, body, &resp)
			case "DELETE":
				err = client.Delete(cmd.Context(), path, &resp)
			default:
				return fmt.Errorf("unsupported method: %s", method)
			}

			if err != nil {
				return err
			}

			// Raw API always outputs JSON (no table rendering).
			f := output.New("json")
			return f.Format(a.Out, resp)
		},
	}

	cmd.Flags().StringVarP(&data, "data", "d", "", "JSON request body")
	return cmd
}
