package main

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var targetRegex = regexp.MustCompile("^([^:]*)(?::(.+))?$")

type basicAuth struct {
	username string
	password string
}

type request struct {
	host      string
	port      int
	path      string
	basicAuth *basicAuth
}

func parseRequest(url *url.URL) (*request, error) {
	parts := strings.Split(url.Path, "/")
	matches := targetRegex.FindStringSubmatch(parts[1])

	if len(matches) != 3 {
		return nil, errors.New("Invalid URL")
	}

	request := request{
		host: matches[1],
		port: 80,
		path: parts[2],
	}

	if matches[2] != "" {
		port, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, errors.New("Invalid target port")
		}
		request.port = port
	}

	if basicAuthParam := url.Query().Get("basic_auth"); basicAuthParam != "" {
		userAndPwd := strings.SplitN(basicAuthParam, ":", 2)
		// TODO: check array bounds
		request.basicAuth = &basicAuth{username: userAndPwd[0], password: userAndPwd[1]}
	}

	return &request, nil
}
