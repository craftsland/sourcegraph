// Package originmap maps Sourcegraph repository URIs to repository
// origins (i.e., clone URLs). It accepts external customization via
// the ORIGIN_MAP environment variable.
//
// It always includes the mapping
// "github.com/!https://github.com/%.git" (github.com ->
// https://github.com/%.git)
package server

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

type prefixAndOrgin struct {
	Prefix string
	Origin string
}

func init() {
	// Setup original maps initially. Failure here is fatal.
	if err := originMaps.setup(); err != nil {
		log.Fatal(err)
	}
	conf.Watch(func() {
		// Setup origin maps in response to config changes. Failure here should
		// just be logged.
		if err := originMaps.setup(); err != nil {
			log.Println("error setting up origin maps", err)
		}
	})
}

var originMaps = &originMapsT{}

type originMapsT struct {
	// Protects everything below.
	mu sync.RWMutex

	originMap       []prefixAndOrgin
	gitoliteHostMap []prefixAndOrgin

	// reposListOriginMap is a mapping from repo URI (path) to repo origin (clone URL).
	reposListOriginMap map[string]string
}

func (o *originMapsT) getOriginMap() []prefixAndOrgin {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.originMap
}

func (o *originMapsT) getGitoliteHostMap() []prefixAndOrgin {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.gitoliteHostMap
}

func (o *originMapsT) getReposListOriginMap() map[string]string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.reposListOriginMap
}

func (o *originMapsT) setup() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Clear the map values.
	o.originMap = nil
	o.gitoliteHostMap = nil
	o.reposListOriginMap = make(map[string]string)

	var err error
	o.originMap, err = parse(conf.Get().GitOriginMap, 1)
	if err != nil {
		return err
	}

	o.gitoliteHostMap, err = parse(conf.Get().GitoliteHosts, 0)
	if err != nil {
		return err
	}
	for _, entry := range o.gitoliteHostMap {
		o.originMap = append(o.originMap, prefixAndOrgin{Prefix: entry.Prefix, Origin: entry.Origin + ":%"})
	}

	// Add origin map for repos.list configuration.
	for _, c := range conf.Get().ReposList {
		o.reposListOriginMap[c.Path] = c.Url
	}

	// Add origin map for GitHub Enterprise instances of the form "${HOSTNAME}/!git@${HOSTNAME}:%.git"
	for _, c := range conf.Get().Github {
		ghURL, err := url.Parse(c.Url)
		if err != nil {
			return err
		}
		// Clone via SSH if this GitHub Enterprise has a self-signed certificate provided.
		// Otherwise git will run into issues with cloning over HTTPS using an invalid certificate.
		if c.Certificate != "" {
			o.originMap = append(o.originMap, prefixAndOrgin{Prefix: ghURL.Hostname() + "/", Origin: fmt.Sprintf("git@%s:%%.git", ghURL.Hostname())})
		} else {
			var auth string
			if c.Token != "" {
				auth = c.Token + "@"
			}
			o.originMap = append(o.originMap, prefixAndOrgin{Prefix: ghURL.Hostname() + "/", Origin: fmt.Sprintf("%s://%s%s/%%.git", ghURL.Scheme, auth, ghURL.Hostname())})
		}
	}

	o.addGitHubDefaults()
	return nil
}

func (o *originMapsT) addGitHubDefaults() {
	// Note: We add several variants here specifically for reverse, so that if
	// a user-provided clone URL is passed to reverse, it still functions as
	// expected. For the case of OriginMap, the first one is returned (i.e. the
	// order below matters).
	o.originMap = append(o.originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "https://github.com/%.git"})
	o.originMap = append(o.originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "http://github.com/%.git"})
	o.originMap = append(o.originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "ssh://git@github.com:%.git"})
	o.originMap = append(o.originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "git://git@github.com:%.git"})
	o.originMap = append(o.originMap, prefixAndOrgin{Prefix: "github.com/", Origin: "git@github.com:%.git"})

	o.originMap = append(o.originMap, prefixAndOrgin{Prefix: "bitbucket.org/", Origin: "https://bitbucket.org/%.git"})
	o.originMap = append(o.originMap, prefixAndOrgin{Prefix: "bitbucket.org/", Origin: "git@bitbucket.org:%.git"})
}

// OriginMap maps the repo URI to the repository origin (clone URL). Returns empty string if no mapping was found.
func OriginMap(repoURI api.RepoURI) string {
	if origin, ok := originMaps.getReposListOriginMap()[string(repoURI)]; ok {
		return origin
	}
	for _, entry := range originMaps.getOriginMap() {
		if strings.HasPrefix(string(repoURI), entry.Prefix) {
			return strings.Replace(entry.Origin, "%", strings.TrimPrefix(string(repoURI), entry.Prefix), 1)
		}
	}
	return ""
}

func parse(raw string, placeholderCount int) (originMap []prefixAndOrgin, err error) {
	for _, e := range strings.Fields(raw) {
		p := strings.Split(e, "!")
		if len(p) != 2 {
			return nil, fmt.Errorf("invalid entry: %s", e)
		}
		if strings.Count(p[1], "%") != placeholderCount {
			return nil, fmt.Errorf("invalid entry: %s", e)
		}
		originMap = append(originMap, prefixAndOrgin{Prefix: p[0], Origin: p[1]})
	}
	return
}
