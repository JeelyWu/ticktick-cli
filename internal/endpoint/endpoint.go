package endpoint

import "fmt"

type Region string

const (
	RegionTickTick Region = "ticktick"
	RegionDida365  Region = "dida365"
)

type Endpoints struct {
	Region       Region
	AuthorizeURL string
	TokenURL     string
	APIBaseURL   string
}

func ForRegion(raw string) (Endpoints, error) {
	switch Region(raw) {
	case "", RegionTickTick:
		return Endpoints{
			Region:       RegionTickTick,
			AuthorizeURL: "https://ticktick.com/oauth/authorize",
			TokenURL:     "https://ticktick.com/oauth/token",
			APIBaseURL:   "https://api.ticktick.com",
		}, nil
	case RegionDida365:
		return Endpoints{
			Region:       RegionDida365,
			AuthorizeURL: "https://dida365.com/oauth/authorize",
			TokenURL:     "https://dida365.com/oauth/token",
			APIBaseURL:   "https://api.dida365.com",
		}, nil
	default:
		return Endpoints{}, fmt.Errorf("unsupported service region %q", raw)
	}
}
