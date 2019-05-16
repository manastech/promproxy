package main

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"promproxy/resolver"
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
	resolver  resolver.Resolver
	basicAuth *basicAuth
	headers   http.Header
}

func parseRequest(ctx context.Context, url *url.URL) (*request, error) {
	parts := strings.Split(url.Path, "/")
	matches := targetRegex.FindStringSubmatch(parts[1])

	if len(matches) != 3 {
		return nil, errors.New("Invalid URL")
	}

	request := request{
		host:    matches[1],
		port:    80,
		path:    parts[2],
		headers: http.Header{},
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

	for _, reqHeader := range url.Query()["header"] {
		keyAndValue := strings.SplitN(reqHeader, ":", 2)
		if len(keyAndValue) == 2 {
			request.headers.Add(keyAndValue[0], keyAndValue[1])
		}
	}

	// var r resolver.Resolver
	switch url.Query().Get("lookup") {
	case "dns":
		request.resolver = resolver.NewDNSResolver()

	case "docker":
		r, err := resolver.NewDockerResolver(ctx)
		if err != nil {
			return nil, err
		}
		request.resolver = r

	default:
		request.resolver = resolver.NewDNSResolver()
	}

	return &request, nil
}
