// Contains a wrapper around github.com/ctfer-io/go-ctfd.
//
// It injects spans for all API operations, with improved consistency on typings based
// upon internal assumptions (e.g. IDs are all strings in TF while CTFd has integers,
// so integers -> string is safe as a the first is a subset of the second, but
// also string -> integer as CTFd dictates IDs while the user has no power over it).

package provider

import (
	"context"
	"net/http"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/v2/provider/utils"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var apiTransport = api.WithTransport(otelhttp.NewTransport(http.DefaultTransport))

func options(ctx context.Context) []api.Option {
	return []api.Option{
		api.WithContext(ctx),
		apiTransport,
	}
}

func GetNonceAndSession(ctx context.Context, url string) (nonce, session string, err error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return api.GetNonceAndSession(url, options(ctx)...)
}

type Client struct {
	sub *api.Client
}

func NewClient(url, nonce, session, apiKey string) *Client {
	return &Client{
		sub: api.NewClient(url, nonce, session, apiKey),
	}
}

func (cli *Client) Login(ctx context.Context, params *api.LoginParams) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.Login(params, options(ctx)...)
}

// region brackets

func (cli *Client) GetBrackets(ctx context.Context, params *api.GetBracketsParams) ([]*api.Bracket, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetBrackets(params, options(ctx)...)
}

func (cli *Client) PostBrackets(ctx context.Context, params *api.PostBracketsParams) (*api.Bracket, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostBrackets(params, options(ctx)...)
}

func (cli *Client) PatchBrackets(ctx context.Context, id string, params *api.PatchBracketsParams) (*api.Bracket, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchBrackets(utils.Atoi(id), params, options(ctx)...)
}

func (cli *Client) DeleteBrackets(ctx context.Context, id string) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteBrackets(utils.Atoi(id), options(ctx)...)
}

// region challenges

func (cli *Client) GetChallenges(ctx context.Context, params *api.GetChallengesParams) ([]*api.Challenge, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetChallenges(params, options(ctx)...)
}

func (cli *Client) PostChallenges(ctx context.Context, params *api.PostChallengesParams) (*api.Challenge, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostChallenges(params, options(ctx)...)
}

func (cli *Client) PatchChallenge(ctx context.Context, id string, params *api.PatchChallengeParams) (*api.Challenge, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchChallenge(utils.Atoi(id), params, options(ctx)...)
}

func (cli *Client) GetChallengeTags(ctx context.Context, id string) ([]*api.Tag, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetChallengeTags(utils.Atoi(id), options(ctx)...)
}

func (cli *Client) GetChallengeRequirements(ctx context.Context, id string) (*api.Requirements, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetChallengeRequirements(utils.Atoi(id), options(ctx)...)
}

func (cli *Client) GetChallengeFiles(ctx context.Context, id string) ([]*api.File, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetChallengeFiles(utils.Atoi(id), options(ctx)...)
}

func (cli *Client) GetChallengeHints(ctx context.Context, id string) ([]*api.Hint, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetChallengeHints(utils.Atoi(id), options(ctx)...)
}

// region tags

func (cli *Client) PostTags(ctx context.Context, params *api.PostTagsParams) (*api.Tag, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostTags(params, options(ctx)...)
}

func (cli *Client) DeleteTag(ctx context.Context, id string) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteTag(id, options(ctx)...)
}

func (cli *Client) DeleteChallenge(ctx context.Context, id string) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteChallenge(utils.Atoi(id), options(ctx)...)
}

func (cli *Client) GetChallenge(ctx context.Context, id string) (*api.Challenge, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetChallenge(utils.Atoi(id), options(ctx)...)
}

// region topics

func (cli *Client) PostTopics(ctx context.Context, params *api.PostTopicsParams) (*api.Topic, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostTopics(params, options(ctx)...)
}

func (cli *Client) DeleteTopic(ctx context.Context, params *api.DeleteTopicArgs) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteTopic(params, options(ctx)...)
}

func (cli *Client) GetChallengeTopics(ctx context.Context, id string) ([]*api.Topic, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetChallengeTopics(utils.Atoi(id), options(ctx)...)
}

// region files

func (cli *Client) PostFiles(ctx context.Context, params *api.PostFilesParams) ([]*api.File, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostFiles(params, options(ctx)...)
}

func (cli *Client) GetFile(ctx context.Context, id string) (*api.File, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetFile(id, options(ctx)...)
}

func (cli *Client) GetFileContent(ctx context.Context, file *api.File) ([]byte, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetFileContent(file, options(ctx)...)
}

func (cli *Client) DeleteFile(ctx context.Context, id string) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteFile(id, options(ctx)...)
}

// region flags

func (cli *Client) PostFlags(ctx context.Context, params *api.PostFlagsParams) (*api.Flag, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostFlags(params, options(ctx)...)
}

func (cli *Client) GetFlag(ctx context.Context, id string) (*api.Flag, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetFlag(id, options(ctx)...)
}

func (cli *Client) PatchFlag(ctx context.Context, id string, params *api.PatchFlagParams) (*api.Flag, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchFlag(id, params, options(ctx)...)
}

func (cli *Client) DeleteFlag(ctx context.Context, id string) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteFlag(id, options(ctx)...)
}

// region hints

func (cli *Client) PostHints(ctx context.Context, params *api.PostHintsParams) (*api.Hint, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostHints(params, options(ctx)...)
}

func (cli *Client) GetHint(ctx context.Context, id string, params *api.GetHintParams) (*api.Hint, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetHint(id, params, options(ctx)...)
}

func (cli *Client) PatchHint(ctx context.Context, id string, params *api.PatchHintsParams) (*api.Hint, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchHint(id, params, options(ctx)...)
}

func (cli *Client) DeleteHint(ctx context.Context, id string) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteHint(id, options(ctx)...)
}

// region solutions

func (cli *Client) PostSolutions(ctx context.Context, params *api.PostSolutionsParams) (*api.Solution, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostSolutions(params, options(ctx)...)
}

func (cli *Client) GetSolutions(ctx context.Context, id string, params *api.GetSolutionsParams) (*api.Solution, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetSolutions(utils.Atoi(id), params, options(ctx)...)
}

func (cli *Client) PatchSolutions(ctx context.Context, id string, params *api.PatchSolutionsParams) (*api.Solution, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchSolutions(utils.Atoi(id), params, options(ctx)...)
}

func (cli *Client) DeleteSolutions(ctx context.Context, id string) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteSolutions(utils.Atoi(id), options(ctx)...)
}

// region teams

func (cli *Client) GetTeams(ctx context.Context, params *api.GetTeamsParams) ([]*api.Team, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetTeams(params, options(ctx)...)
}

func (cli *Client) PostTeams(ctx context.Context, params *api.PostTeamsParams) (*api.Team, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostTeams(params, options(ctx)...)
}

func (cli *Client) PatchTeam(ctx context.Context, id string, params *api.PatchTeamsParams) (*api.Team, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchTeam(utils.Atoi(id), params, options(ctx)...)
}

func (cli *Client) GetTeam(ctx context.Context, id string) (*api.Team, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetTeam(utils.Atoi(id), options(ctx)...)
}

func (cli *Client) DeleteTeam(ctx context.Context, id string) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteTeam(utils.Atoi(id), options(ctx)...)
}

func (cli *Client) PostTeamMembers(ctx context.Context, id string, params *api.PostTeamsMembersParams) (int, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostTeamMembers(utils.Atoi(id), params, options(ctx)...)
}

func (cli *Client) GetTeamMembers(ctx context.Context, id string) ([]int, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetTeamMembers(utils.Atoi(id), options(ctx)...)
}

func (cli *Client) DeleteTeamMembers(ctx context.Context, id string, params *api.DeleteTeamMembersParams) ([]int, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteTeamMembers(utils.Atoi(id), params, options(ctx)...)
}

// region users

func (cli *Client) GetUsers(ctx context.Context, params *api.GetUsersParams) ([]*api.User, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetUsers(params, options(ctx)...)
}

func (cli *Client) PostUsers(ctx context.Context, params *api.PostUsersParams) (*api.User, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PostUsers(params, options(ctx)...)
}

func (cli *Client) GetUser(ctx context.Context, id string) (*api.User, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.GetUser(utils.Atoi(id), options(ctx)...)
}

func (cli *Client) PatchUser(ctx context.Context, id string, params *api.PatchUsersParams) (*api.User, error) {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.PatchUser(utils.Atoi(id), params, options(ctx)...)
}

func (cli *Client) DeleteUser(ctx context.Context, id string) error {
	ctx, span := StartAPISpan(ctx)
	defer span.End()

	return cli.sub.DeleteUser(utils.Atoi(id), options(ctx)...)
}
