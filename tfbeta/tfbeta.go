package tfbeta

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

// TFBeta is TestFlight beta.
// It helps to check whether beta is full.
// Also, TFBeta helps to get an app name that beta belongs.
type TFBeta struct {
	link    string
	appName string
}

// NewTFBeta returns new TestFlight beta.
// ErrInvalidTestFlightLink can be returned if link is invalid.
func NewTFBeta(link string) (*TFBeta, error) {
	if !isValid(link) {
		return nil, ErrInvalidTestFlightLink
	}

	appName, err := extractAppName(link)
	if err != nil {
		return nil, ErrInvalidTestFlightLink
	}

	return &TFBeta{link: link, appName: appName}, nil
}

// IsFull returns whether beta is full or not.
func (r *TFBeta) IsFull() (bool, error) {
	resp, err := http.Get(r.link)
	if err != nil {
		return false, fmt.Errorf("%s: %w", ErrUnexpected, err)
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

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
func (r *TFBeta) GetAppName() string {
	return r.appName
}

// GetLink returns beta link.
func (r *TFBeta) GetLink() string {
	return r.link
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
