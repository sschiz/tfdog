package beta

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	"golang.org/x/net/html"
)

const (
	testFlightHost = "testflight.apple.com"
	pathLen        = 14
	joinPathPart   = "/join/"
)

var (
	// ErrInvalidTestFlightLink is error that will be returned
	// if TestFlight link is invalid.
	ErrInvalidTestFlightLink = errors.New("invalid TestFlight link")

	// ErrUnexpected is error that will be returned
	// when unexpected case is happened.
	ErrUnexpected = errors.New("unexpected error")

	// ErrStatusNotOK is error that will be returned
	// if HTTP status is not 200.
	ErrStatusNotOK = errors.New("HTTP status is not 200")
)
var re = regexp.MustCompile(`the (.*) beta`)

// Beta is TestFlight beta.
// It helps to check whether beta is full.
// Also, Beta helps to get an app name that beta belongs.
type Beta struct {
	link    string
	appName string

	client *http.Client
	req    *http.Request
}

// NewTFBeta returns new TestFlight beta.
// ErrInvalidTestFlightLink can be returned if link is invalid.
func NewTFBeta(link string) (*Beta, error) {
	if !isValid(link) {
		return nil, ErrInvalidTestFlightLink
	}

	appName, err := extractAppName(link)
	if err != nil {
		return nil, ErrInvalidTestFlightLink
	}

	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		panic(err)
	}

	return &Beta{
		link:    link,
		appName: appName,
		client:  http.DefaultClient,
		req:     req,
	}, nil
}

// IsFull returns whether beta is full or not.
func (r *Beta) IsFull() (bool, error) {
	resp, err := r.client.Do(r.req)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, ErrStatusNotOK
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("%s: %w", ErrUnexpected, err)
	}

	isFull := bytes.Contains(body, []byte("This beta is full."))
	return isFull, nil
}

// GetAppName returns app name that beta belongs.
func (r *Beta) GetAppName() string {
	return r.appName
}

// GetLink returns beta link.
func (r *Beta) GetLink() string {
	return r.link
}

// WithClient overrides HTTP client.
func (r *Beta) WithClient(client *http.Client) *Beta {
	r.client = client
	return r
}

func isValid(link string) bool {
	u, err := url.Parse(link)
	if err != nil {
		return false
	}

	if u.Host != testFlightHost || len(u.Path) != pathLen || u.Path[0:6] != joinPathPart {
		return false
	}

	return true
}

func extractAppName(link string) (string, error) {
	resp, err := http.Get(link)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrUnexpected, err)
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", ErrStatusNotOK
	}

	z := html.NewTokenizer(resp.Body)
	var (
		isTitle bool
		title   []byte
	)

	const titleTag = "title"
cl:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			err = z.Err()
			if err != io.EOF {
				return "", err
			}
			break cl
		case html.StartTagToken, html.EndTagToken:
			if z.Token().Data == titleTag {
				isTitle = true
			}
		case html.TextToken:
			if isTitle {
				title = z.Text()
				break cl
			}
		}
	}

	match := re.FindSubmatch(title)
	if len(match) < 2 {
		panic("unexpected to not submatch")
	}

	return string(match[1]), nil
}
