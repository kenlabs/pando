package command

import (
	"fmt"
	"github.com/kenlabs/pando/cmd/client/command/admin"
	"github.com/kenlabs/pando/cmd/client/command/metadata"
	"github.com/kenlabs/pando/cmd/client/command/pando"
	"github.com/kenlabs/pando/cmd/client/command/provider"
	"github.com/spf13/cobra"
	"net/url"

	"github.com/kenlabs/pando/cmd/client/command/api"
)

func NewRoot() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:        "pando",
		Short:      "Pando client cli",
		SuggestFor: []string{"pando"},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			_, err := url.Parse(api.PandoAPIBaseURL)
			if api.PandoAPIBaseURL == "" || err != nil {
				return fmt.Errorf("pando api url is invalid, given: \"%s\"\n", api.PandoAPIBaseURL)
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&api.PandoAPIBaseURL, "pando-api", "a", "",
		"set pando api url")
	api.NewClient(api.PandoAPIBaseURL)

	childCommands := []*cobra.Command{
		provider.NewProviderCmd(),
		metadata.NewMetadataCmd(),
		pando.NewPandoCmd(),
		admin.NewAdminCmd(),
	}
	rootCmd.AddCommand(childCommands...)

	return rootCmd
}
