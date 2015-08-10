package api

import (
	"strconv"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Repo interface {
	Save(*Todo) error
	Get(string) (*Todo, error)
	Query(*Meta) (TodoCollection, error)
}

func NewGhRepo(user string, repo string, token string) *GhRepo {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	return &GhRepo{
		Client: client,
		Owner:  user,
		Repo:   repo,
	}
}

type GhRepo struct {
	Client *github.Client
	Owner  string
	Repo   string
}

func (r *GhRepo) Save(todo *Todo) error {
	if todo.Id == "" {
		labels := append(todo.Meta.Tags, todo.Meta.Context)

		req := &github.IssueRequest{
			Title:  &todo.Subject,
			Labels: &labels,
		}
		issue, _, err := r.Client.Issues.Create(r.Owner, r.Repo, req)
		if err != nil {
			return err
		}
		todo.Id = strconv.Itoa(*issue.Number)
	} else {
		idInt, err := strconv.Atoi(todo.Id)
		if err != nil {
			return err
		}
		labels := append(todo.Meta.Tags, todo.Meta.Context)

		req := &github.IssueRequest{
			Title:  &todo.Subject,
			Labels: &labels,
			State:  &todo.Status,
		}
		_, _, err = r.Client.Issues.Edit(r.Owner, r.Repo, idInt, req)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *GhRepo) Get(id string) (*Todo, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}

	issue, _, err := r.Client.Issues.Get(r.Owner, r.Repo, idInt)
	if err != nil {
		return nil, err
	}

	todo := parseTodoFromIssue(*issue)
	return todo, nil
}

func (r *GhRepo) Query(meta *Meta) (TodoCollection, error) {
	todos := make([]*Todo, 0)

	labelStrings := append(meta.Tags, meta.Context)
	labels := &github.IssueListByRepoOptions{
		Labels: labelStrings,
	}
	issues, _, err := r.Client.Issues.ListByRepo(r.Owner, r.Repo, labels)
	if err != nil {
		return nil, err
	}

	for _, issue := range issues {
		todos = append(todos, parseTodoFromIssue(issue))
	}

	var c TodoCollection = todos
	return c, nil
}
func parseTodoFromIssue(issue github.Issue) *Todo {
	meta := parseMetaFromIssue(issue)

	todo := &Todo{
		Id:      strconv.Itoa(*issue.Number),
		Subject: *issue.Title,
		Status:  *issue.State,
		Meta:    meta,
	}
	return todo
}
func parseMetaFromIssue(issue github.Issue) *Meta {
	meta := &Meta{}
	for _, label := range issue.Labels {
		name := *label.Name
		if name[0:1] == ContextPrefix {
			meta.Context = name
		}
		if name[0:1] == TagPrefix {
			meta.Tags = append(meta.Tags, name)
		}
	}
	return meta
}
