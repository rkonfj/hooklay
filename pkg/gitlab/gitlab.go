package gitlab

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/rkonfj/opkit/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

func Gitlab() *gitlab.Client {
	git, err := gitlab.NewClient(config.Conf.Gitlab.Token, gitlab.WithBaseURL(config.Conf.Gitlab.BaseURL))
	if err != nil {
		logrus.Fatal(err)
	}
	return git
}

func Changelog(oldCommitID, newCommitID string, projectId int) string {
	git := Gitlab()

	compare, _, err := git.Repositories.Compare(projectId, &gitlab.CompareOptions{From: &oldCommitID, To: &newCommitID})
	if err != nil {
		logrus.Error(err)
		return err.Error()
	}
	var changelog bytes.Buffer
	for _, commit := range compare.Commits {
		if strings.Contains(commit.Message, "Merge") {
			continue
		}
		fmt.Fprintf(&changelog, "   - %s %s %s", commit.CommittedDate, commit.AuthorEmail, commit.Message)
	}
	return changelog.String()
}
