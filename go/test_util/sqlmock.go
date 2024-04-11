package test_util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type sqlQueryMatcherImpl struct {
	whitespaceRegex *regexp.Regexp
}

func newSQLQueryMatcher() sqlmock.QueryMatcher {
	return &sqlQueryMatcherImpl{
		whitespaceRegex: regexp.MustCompile(`\s+`),
	}
}

// Match implements sqlmock.QueryMatcher.
func (s *sqlQueryMatcherImpl) Match(expected string, actual string) error {
	expected = strings.TrimSpace(s.whitespaceRegex.ReplaceAllString(expected, " "))
	actual = strings.TrimSpace(s.whitespaceRegex.ReplaceAllString(actual, " "))
	if expected == actual {
		return nil
	}
	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(expected, actual, false)
	return fmt.Errorf("expected and actual SQL query do not match: %s", dmp.DiffPrettyHtml(diffs))
}

var _ sqlmock.QueryMatcher = (*sqlQueryMatcherImpl)(nil)

var SQLQueryMatcher = newSQLQueryMatcher()
