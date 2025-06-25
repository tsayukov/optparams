// This file is licensed under the terms of the MIT License (see LICENSE file)
// Copyright (c) 2025 Pavel Tsayukov p.tsayukov@gmail.com

package optparams_test

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/tsayukov/optparams"
)

// Client is an example of a struct that contains fields, some of which
// may be required and others optional, with or without default values.
type Client struct {
	httpClient *http.Client
	endpoint   string
	apiToken   string
}

const DefaultEndpoint = "https://github.com/tsayukov/optparams"

func NewClient(apiToken string, opts ...optparams.Func[Client]) (*Client, error) {
	// Create a new struct with initialized fields by the required arguments.
	c := &Client{
		apiToken: apiToken,
	}

	// Append default values for optional parameters, if any.
	opts = append(opts,
		optparams.Default[Client](&c.httpClient, http.DefaultClient),
		optparams.Default[Client](&c.endpoint, DefaultEndpoint),
	)

	// Apply all the optional arguments one by one to the created struct.
	if err := optparams.Apply(c, opts...); err != nil {
		return nil, err
	}

	return c, nil
}

// WithHttpClient is an example of a function to pass an optional argument.
func WithHttpClient(httpClient *http.Client) optparams.Func[Client] {
	return func(c *Client) error {
		if httpClient == nil {
			return optparams.ErrFailFast
		}

		c.httpClient = httpClient

		return nil
	}
}

// WithEndpoint is an example of a function to pass an optional argument.
func WithEndpoint(url string) optparams.Func[Client] {
	return func(c *Client) error {
		if url == "" {
			return errors.New("endpoint is empty")
		}

		c.endpoint = url

		return nil
	}
}

// GetAPIToken imitates getting a secret API token from the environment.
func GetAPIToken(name string) (string, error) {
	if err := os.Setenv(name, rand.Text()); err != nil {
		return "", err
	}

	token, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("%s environment variable is not set", name)
	}

	return token, nil
}

func Example_struct() {
	apiToken, err := GetAPIToken("API_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	httpClient := &http.Client{Timeout: time.Minute}

	client, err := NewClient(
		// Required arguments:
		apiToken,
		// Optional arguments:
		WithHttpClient(httpClient),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reflect.DeepEqual(*client, Client{
		httpClient: httpClient,
		endpoint:   DefaultEndpoint,
		apiToken:   apiToken,
	}))

	// Output: true
}
