package compute

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Output struct {
	MatchPattern  MatchPattern
	OutputPattern string
	Separator     string
}

func (c *Output) String() string {
	return fmt.Sprintf("Output with separator: (%s) -> (%s) separator: %s", c.MatchPattern.String(), c.OutputPattern, c.Separator)
}

func substituteRegexp(content string, match *regexp.Regexp, replacePattern, separator string) string {
	result := []byte{}
	for _, submatches := range match.FindAllStringSubmatchIndex(content, -1) {
		result = append(match.ExpandString(result, replacePattern, content, submatches), separator...)
	}
	return string(result)
}

func output(ctx context.Context, fragment []byte, matchPattern MatchPattern, replacePattern string, separator string) (*Text, error) {
	var newFragment string
	switch match := matchPattern.(type) {
	case *Regexp:
		newFragment = substituteRegexp(string(fragment), match.Value, replacePattern, separator)
	case *Comby:
		return nil, nil
	}
	return &Text{Value: newFragment, Kind: "output"}, nil
}

func (c *Output) Run(ctx context.Context, fm *result.FileMatch) (Result, error) {
	var lines []string
	for _, line := range fm.LineMatches {
		lines = append(lines, line.Preview)
	}
	fragment := strings.Join(lines, "\n")
	return output(ctx, []byte(fragment), c.MatchPattern, c.OutputPattern, c.Separator)
}
