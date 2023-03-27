package operators

import (
	"bruce/loader"
	"bruce/rest"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"strings"
	"time"
)

type GHRelease struct {
	URL             string     `json:"url"`
	AssetsURL       string     `json:"assets_url"`
	UploadURL       string     `json:"upload_url"`
	HTMLURL         string     `json:"html_url"`
	ID              int        `json:"id"`
	Author          GHAuthor   `json:"author"`
	NodeID          string     `json:"node_id"`
	TagName         string     `json:"tag_name"`
	TargetCommitish string     `json:"target_commitish"`
	Name            string     `json:"name"`
	Draft           bool       `json:"draft"`
	Prerelease      bool       `json:"prerelease"`
	CreatedAt       time.Time  `json:"created_at"`
	PublishedAt     time.Time  `json:"published_at"`
	Assets          []GHAssets `json:"assets"`
	TarballURL      string     `json:"tarball_url"`
	ZipballURL      string     `json:"zipball_url"`
	Body            string     `json:"body"`
}
type GHAuthor struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}
type GHUploader struct {
	Login             string `json:"login"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	AvatarURL         string `json:"avatar_url"`
	GravatarID        string `json:"gravatar_id"`
	URL               string `json:"url"`
	HTMLURL           string `json:"html_url"`
	FollowersURL      string `json:"followers_url"`
	FollowingURL      string `json:"following_url"`
	GistsURL          string `json:"gists_url"`
	StarredURL        string `json:"starred_url"`
	SubscriptionsURL  string `json:"subscriptions_url"`
	OrganizationsURL  string `json:"organizations_url"`
	ReposURL          string `json:"repos_url"`
	EventsURL         string `json:"events_url"`
	ReceivedEventsURL string `json:"received_events_url"`
	Type              string `json:"type"`
	SiteAdmin         bool   `json:"site_admin"`
}
type GHAssets struct {
	URL                string     `json:"url"`
	ID                 int        `json:"id"`
	NodeID             string     `json:"node_id"`
	Name               string     `json:"name"`
	Label              string     `json:"label"`
	Uploader           GHUploader `json:"uploader"`
	ContentType        string     `json:"content_type"`
	State              string     `json:"state"`
	Size               int        `json:"size"`
	DownloadCount      int        `json:"download_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	BrowserDownloadURL string     `json:"browser_download_url"`
}

// Github provides a means to download github release files.
type Github struct {
	Repo       string `yaml:"githubRepo"`
	Version    string `yaml:"releaseVer"`
	Asset      string `yaml:"assetType"`
	AssetMatch string `yaml:"strContains"`
	Storage    string `yaml:"localDir"`
	DoExtract  bool   `yaml:"doExtract"`
	StripRoot  bool   `yaml:"stripRoot"`
	Owner      string
	RepoName   string
	GHClient   *rest.RESTClient
}

func (g *Github) Execute() error {

	preStrip := strings.TrimLeft(g.Repo, "https://")
	strip := strings.TrimRight(preStrip, ".git")
	parts := strings.Split(strip, "/")
	g.Owner = parts[1]
	g.RepoName = parts[2]
	var gr GHRelease
	gr, err := g.getRelease()
	if err != nil {
		return err
	}
	url := g.getUrlFromType(gr)
	if g.DoExtract {

	}
	rf, fn, err := loader.ReadRemoteFile(url)
	if err != nil {
		log.Error().Err(err).Msg("could not read remote github file")
		return err
	}
	localFile := path.Join(g.Storage, fn)
	log.Info().Msgf("wrote %s from github release", localFile)
	return os.WriteFile(localFile, rf, 0664)
}

func (g *Github) getRelease() (GHRelease, error) {
	if g.Version != "latest" {
		return g.getReleaseMatch()
	}
	return g.getLatestRel()
}

func (g *Github) getClient() (*rest.RESTClient, error) {
	if g.GHClient != nil {
		return g.GHClient, nil
	}
	ghr, err := rest.NewRestClient("https://api.github.com", false)
	if err != nil {
		log.Error().Err(err).Msg("github client is broken")
		return nil, err
	}
	g.GHClient = ghr
	return ghr, nil
}

func (g *Github) getReleaseMatch() (GHRelease, error) {
	c, err := g.getClient()
	if err != nil {
		return GHRelease{}, err
	}
	var releases []GHRelease
	err = c.Get(fmt.Sprintf("/%s/%s/releases", g.Owner, g.RepoName), nil, releases)
	if err != nil {
		return GHRelease{}, err
	}
	for _, v := range releases {
		if v.Name == g.Version {
			return v, nil

		}
	}
	return GHRelease{}, fmt.Errorf("no such version for repostorty: github.com/%s/%s", g.Owner, g.RepoName)
}
func (g *Github) getLatestRel() (GHRelease, error) {
	c, err := g.getClient()
	if err != nil {
		return GHRelease{}, err
	}
	var release GHRelease
	err = c.Get(fmt.Sprintf("/%s/%s/releases", g.Owner, g.RepoName), nil, release)
	if err != nil {
		return GHRelease{}, err
	}
	return release, nil
}

func (g *Github) getUrlFromType(rel GHRelease) string {
	if g.Asset == "zipball" {
		return rel.ZipballURL
	}
	if g.Asset == "tarball" {
		return rel.TarballURL
	}
	// if it does have an asset match string loop through and try to find it
	if g.AssetMatch != "" {
		for _, a := range rel.Assets {
			if strings.Contains(a.Name, g.AssetMatch) {
				return a.URL
			}
		}
	}

	// default we just return the repository tarball
	return rel.TarballURL
}
