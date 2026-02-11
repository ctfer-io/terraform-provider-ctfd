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

func apiOptions(ctx context.Context) []api.Option {
	return []api.Option{
		api.WithContext(ctx),
		apiTransport,
	}
}

func GetNonceAndSession(ctx context.Context, url string, opts ...Option) (nonce, session string, err error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return api.GetNonceAndSession(url, apiOptions(ctx)...)
}

type Client struct {
	sub *api.Client
}

func NewClient(url, nonce, session, apiKey string) *Client {
	return &Client{
		sub: api.NewClient(url, nonce, session, apiKey),
	}
}

func (cli *Client) Login(ctx context.Context, params *api.LoginParams, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.Login(params, apiOptions(ctx)...)
}

// region brackets

func (cli *Client) GetBrackets(ctx context.Context, params *api.GetBracketsParams, opts ...Option) ([]*api.Bracket, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetBrackets(params, apiOptions(ctx)...)
}

func (cli *Client) PostBrackets(ctx context.Context, params *api.PostBracketsParams, opts ...Option) (*api.Bracket, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostBrackets(params, apiOptions(ctx)...)
}

func (cli *Client) PatchBrackets(ctx context.Context, id string, params *api.PatchBracketsParams, opts ...Option) (*api.Bracket, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PatchBrackets(utils.Atoi(id), params, apiOptions(ctx)...)
}

func (cli *Client) DeleteBrackets(ctx context.Context, id string, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteBrackets(utils.Atoi(id), apiOptions(ctx)...)
}

// region challenges

func (cli *Client) GetChallenges(ctx context.Context, params *api.GetChallengesParams, opts ...Option) ([]*api.Challenge, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetChallenges(params, apiOptions(ctx)...)
}

func (cli *Client) PostChallenges(ctx context.Context, params *api.PostChallengesParams, opts ...Option) (*api.Challenge, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostChallenges(params, apiOptions(ctx)...)
}

func (cli *Client) PatchChallenge(ctx context.Context, id string, params *api.PatchChallengeParams, opts ...Option) (*api.Challenge, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PatchChallenge(utils.Atoi(id), params, apiOptions(ctx)...)
}

func (cli *Client) GetChallengeTags(ctx context.Context, id string, opts ...Option) ([]*api.Tag, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetChallengeTags(utils.Atoi(id), apiOptions(ctx)...)
}

func (cli *Client) GetChallengeRequirements(ctx context.Context, id string, opts ...Option) (*api.Requirements, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetChallengeRequirements(utils.Atoi(id), apiOptions(ctx)...)
}

func (cli *Client) GetChallengeFiles(ctx context.Context, id string, opts ...Option) ([]*api.File, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetChallengeFiles(utils.Atoi(id), apiOptions(ctx)...)
}

func (cli *Client) GetChallengeHints(ctx context.Context, id string, opts ...Option) ([]*api.Hint, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetChallengeHints(utils.Atoi(id), apiOptions(ctx)...)
}

// region tags

func (cli *Client) PostTags(ctx context.Context, params *api.PostTagsParams, opts ...Option) (*api.Tag, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostTags(params, apiOptions(ctx)...)
}

func (cli *Client) DeleteTag(ctx context.Context, id string, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteTag(id, apiOptions(ctx)...)
}

func (cli *Client) DeleteChallenge(ctx context.Context, id string, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteChallenge(utils.Atoi(id), apiOptions(ctx)...)
}

func (cli *Client) GetChallenge(ctx context.Context, id string, opts ...Option) (*api.Challenge, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetChallenge(utils.Atoi(id), apiOptions(ctx)...)
}

// region topics

func (cli *Client) PostTopics(ctx context.Context, params *api.PostTopicsParams, opts ...Option) (*api.Topic, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostTopics(params, apiOptions(ctx)...)
}

func (cli *Client) DeleteTopic(ctx context.Context, params *api.DeleteTopicArgs, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteTopic(params, apiOptions(ctx)...)
}

func (cli *Client) GetChallengeTopics(ctx context.Context, id string, opts ...Option) ([]*api.Topic, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetChallengeTopics(utils.Atoi(id), apiOptions(ctx)...)
}

// region files

func (cli *Client) PostFiles(ctx context.Context, params *api.PostFilesParams, opts ...Option) ([]*api.File, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostFiles(params, apiOptions(ctx)...)
}

func (cli *Client) GetFile(ctx context.Context, id string, opts ...Option) (*api.File, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetFile(id, apiOptions(ctx)...)
}

func (cli *Client) GetFileContent(ctx context.Context, file *api.File, opts ...Option) ([]byte, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetFileContent(file, apiOptions(ctx)...)
}

func (cli *Client) DeleteFile(ctx context.Context, id string, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteFile(id, apiOptions(ctx)...)
}

// region flags

func (cli *Client) PostFlags(ctx context.Context, params *api.PostFlagsParams, opts ...Option) (*api.Flag, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostFlags(params, apiOptions(ctx)...)
}

func (cli *Client) GetFlag(ctx context.Context, id string, opts ...Option) (*api.Flag, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetFlag(id, apiOptions(ctx)...)
}

func (cli *Client) PatchFlag(ctx context.Context, id string, params *api.PatchFlagParams, opts ...Option) (*api.Flag, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PatchFlag(id, params, apiOptions(ctx)...)
}

func (cli *Client) DeleteFlag(ctx context.Context, id string, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteFlag(id, apiOptions(ctx)...)
}

// region hints

func (cli *Client) PostHints(ctx context.Context, params *api.PostHintsParams, opts ...Option) (*api.Hint, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostHints(params, apiOptions(ctx)...)
}

func (cli *Client) GetHint(ctx context.Context, id string, params *api.GetHintParams, opts ...Option) (*api.Hint, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetHint(id, params, apiOptions(ctx)...)
}

func (cli *Client) PatchHint(ctx context.Context, id string, params *api.PatchHintsParams, opts ...Option) (*api.Hint, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PatchHint(id, params, apiOptions(ctx)...)
}

func (cli *Client) DeleteHint(ctx context.Context, id string, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteHint(id, apiOptions(ctx)...)
}

// region solutions

func (cli *Client) PostSolutions(ctx context.Context, params *api.PostSolutionsParams, opts ...Option) (*api.Solution, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostSolutions(params, apiOptions(ctx)...)
}

func (cli *Client) GetSolutions(ctx context.Context, id string, params *api.GetSolutionsParams, opts ...Option) (*api.Solution, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetSolutions(utils.Atoi(id), params, apiOptions(ctx)...)
}

func (cli *Client) PatchSolutions(ctx context.Context, id string, params *api.PatchSolutionsParams, opts ...Option) (*api.Solution, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PatchSolutions(utils.Atoi(id), params, apiOptions(ctx)...)
}

func (cli *Client) DeleteSolutions(ctx context.Context, id string, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteSolutions(utils.Atoi(id), apiOptions(ctx)...)
}

// region teams

func (cli *Client) GetTeams(ctx context.Context, params *api.GetTeamsParams, opts ...Option) ([]*api.Team, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetTeams(params, apiOptions(ctx)...)
}

func (cli *Client) PostTeams(ctx context.Context, params *api.PostTeamsParams, opts ...Option) (*api.Team, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostTeams(params, apiOptions(ctx)...)
}

func (cli *Client) PatchTeam(ctx context.Context, id string, params *api.PatchTeamsParams, opts ...Option) (*api.Team, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PatchTeam(utils.Atoi(id), params, apiOptions(ctx)...)
}

func (cli *Client) GetTeam(ctx context.Context, id string, opts ...Option) (*api.Team, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetTeam(utils.Atoi(id), apiOptions(ctx)...)
}

func (cli *Client) DeleteTeam(ctx context.Context, id string, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteTeam(utils.Atoi(id), apiOptions(ctx)...)
}

func (cli *Client) PostTeamMembers(ctx context.Context, id string, params *api.PostTeamsMembersParams, opts ...Option) (int, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostTeamMembers(utils.Atoi(id), params, apiOptions(ctx)...)
}

func (cli *Client) GetTeamMembers(ctx context.Context, id string, opts ...Option) ([]int, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetTeamMembers(utils.Atoi(id), apiOptions(ctx)...)
}

func (cli *Client) DeleteTeamMembers(ctx context.Context, id string, params *api.DeleteTeamMembersParams, opts ...Option) ([]int, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteTeamMembers(utils.Atoi(id), params, apiOptions(ctx)...)
}

// region users

func (cli *Client) GetUsers(ctx context.Context, params *api.GetUsersParams, opts ...Option) ([]*api.User, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetUsers(params, apiOptions(ctx)...)
}

func (cli *Client) PostUsers(ctx context.Context, params *api.PostUsersParams, opts ...Option) (*api.User, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PostUsers(params, apiOptions(ctx)...)
}

func (cli *Client) GetUser(ctx context.Context, id string, opts ...Option) (*api.User, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.GetUser(utils.Atoi(id), apiOptions(ctx)...)
}

func (cli *Client) PatchUser(ctx context.Context, id string, params *api.PatchUsersParams, opts ...Option) (*api.User, error) {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.PatchUser(utils.Atoi(id), params, apiOptions(ctx)...)
}

func (cli *Client) DeleteUser(ctx context.Context, id string, opts ...Option) error {
	ctx, span := StartAPISpan(ctx, getTracer(opts...))
	defer span.End()

	return cli.sub.DeleteUser(utils.Atoi(id), apiOptions(ctx)...)
}
