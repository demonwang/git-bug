package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/MichaelMure/git-bug/cache"
	"github.com/MichaelMure/git-bug/identity"
	"github.com/MichaelMure/git-bug/util/colors"
	"github.com/MichaelMure/git-bug/util/interrupt"
)

var (
	userLsOutputFormat string
)

func runUserLs(_ *cobra.Command, _ []string) error {
	backend, err := cache.NewRepoCache(repo)
	if err != nil {
		return err
	}
	defer backend.Close()
	interrupt.RegisterCleaner(backend.Close)

	ids := backend.AllIdentityIds()
	var users []*cache.IdentityExcerpt
	for _, id := range ids {
		user, err := backend.ResolveIdentityExcerpt(id)
		if err != nil {
			return err
		}
		users = append(users, user)
	}

	switch userLsOutputFormat {
	case "json":
		return userLsJsonFormatter(users)
	case "default":
		return userLsDefaultFormatter(users)
	default:
		return fmt.Errorf("unknown format %s", userLsOutputFormat)
	}
}

type JSONIdentity struct {
	Id      string `json:"id"`
	HumanId string `json:"human_id"`
	Name    string `json:"name"`
	Login   string `json:"login"`
}

func NewJSONIdentity(i identity.Interface) JSONIdentity {
	return JSONIdentity{
		Id:      i.Id().String(),
		HumanId: i.Id().Human(),
		Name:    i.Name(),
		Login:   i.Login(),
	}
}

func NewJSONIdentityFromExcerpt(excerpt *cache.IdentityExcerpt) JSONIdentity {
	return JSONIdentity{
		Id:      excerpt.Id.String(),
		HumanId: excerpt.Id.Human(),
		Name:    excerpt.Name,
		Login:   excerpt.Login,
	}
}

func NewJSONIdentityFromLegacyExcerpt(excerpt *cache.LegacyAuthorExcerpt) JSONIdentity {
	return JSONIdentity{
		Name:  excerpt.Name,
		Login: excerpt.Login,
	}
}

func userLsDefaultFormatter(users []*cache.IdentityExcerpt) error {
	for _, user := range users {
		fmt.Printf("%s %s\n",
			colors.Cyan(user.Id.Human()),
			user.DisplayName(),
		)
	}

	return nil
}

func userLsJsonFormatter(users []*cache.IdentityExcerpt) error {
	jsonUsers := make([]JSONIdentity, len(users))
	for i, user := range users {
		jsonUsers[i] = NewJSONIdentityFromExcerpt(user)
	}

	jsonObject, _ := json.MarshalIndent(jsonUsers, "", "    ")
	fmt.Printf("%s\n", jsonObject)
	return nil
}

var userLsCmd = &cobra.Command{
	Use:     "ls",
	Short:   "List identities.",
	PreRunE: loadRepo,
	RunE:    runUserLs,
}

func init() {
	userCmd.AddCommand(userLsCmd)
	userLsCmd.Flags().SortFlags = false
	userLsCmd.Flags().StringVarP(&userLsOutputFormat, "format", "f", "default",
		"Select the output formatting style. Valid values are [default,json]")
}
