package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/niiithish/lazyports/internal/version"
)

const repo = "niiithish/lazyports"

// Info describes whether a newer release is available.
type Info struct {
	Available bool
	Latest    string
	Current   string
	Command   string
}

// Check returns update info by comparing the local version with GitHub's latest tag.
func Check(ctx context.Context) Info {
	current := version.Current()
	info := Info{Current: current, Command: version.InstallCommand}

	if os.Getenv("LAZYPORTS_SKIP_UPDATE") != "" {
		return info
	}

	latest, err := fetchLatestTag(ctx)
	if err != nil || latest == "" {
		return info
	}

	info.Latest = latest
	info.Available = IsNewer(latest, current)
	return info
}

func fetchLatestTag(ctx context.Context) (string, error) {
	if tag, err := fetchReleaseTag(ctx); err == nil && tag != "" {
		return tag, nil
	}
	return fetchFirstTag(ctx)
}

func fetchReleaseTag(ctx context.Context) (string, error) {
	body, err := githubGET(ctx, fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return "", err
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	return normalizeTag(payload.TagName), nil
}

func fetchFirstTag(ctx context.Context) (string, error) {
	body, err := githubGET(ctx, fmt.Sprintf("https://api.github.com/repos/%s/tags?per_page=1", repo))
	if err != nil {
		return "", err
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &tags); err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags")
	}
	return normalizeTag(tags[0].Name), nil
}

func githubGET(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "lazyports/"+version.Current())

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

func normalizeTag(tag string) string {
	return strings.TrimPrefix(strings.TrimSpace(tag), "v")
}

// IsNewer reports whether latest is a higher semver than current.
func IsNewer(latest, current string) bool {
	latest = normalizeTag(latest)
	current = normalizeTag(current)
	if latest == "" || current == "" {
		return false
	}
	if current == "dev" {
		return true
	}
	return Compare(latest, current) > 0
}

// Compare compares semver strings a and b. Returns 1 if a>b, -1 if a<b, 0 if equal.
func Compare(a, b string) int {
	ap := parseSemverParts(a)
	bp := parseSemverParts(b)
	max := len(ap)
	if len(bp) > max {
		max = len(bp)
	}
	for i := 0; i < max; i++ {
		av, bv := 0, 0
		if i < len(ap) {
			av = ap[i]
		}
		if i < len(bp) {
			bv = bp[i]
		}
		switch {
		case av > bv:
			return 1
		case av < bv:
			return -1
		}
	}
	return 0
}

func parseSemverParts(v string) []int {
	chunks := strings.Split(normalizeTag(v), ".")
	out := make([]int, 0, len(chunks))
	for _, chunk := range chunks {
		n := 0
		for _, ch := range chunk {
			if ch < '0' || ch > '9' {
				break
			}
			n = n*10 + int(ch-'0')
		}
		out = append(out, n)
	}
	return out
}
