# optparams

A helper package to add optional parameters to your functions.

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)

## Installation

```bash
go get github.com/tsayukov/optparams
```

## Usage

```go
package main

import (
    "net/http"
    "time"

    "github.com/tsayukov/optparams"
)

type Client struct {
    httpClient *http.Client
    basicURL   string
    apiToken   string
}

func WithHttpClient(httpClient *http.Client) optparams.Func[Client] {
    return func(c *Client) error {
        if httpClient == nil {
            return optparams.ErrFailFast
        }

        c.httpClient = httpClient

        return nil
    }
}

func NewClient(basicURL, apiToken string, opts ...optparams.Func[Client]) (*Client, error) {
    c := &Client{
        basicURL: basicURL, 
        apiToken: apiToken,
    }

    if err := optparams.Apply(c, opts...); err != nil {
        return nil, err
    }

    if c.httpClient == nil {
        c.httpClient = http.DefaultClient
    }

    return c, nil
}

func main() {
    client, err := NewClient(
        "<URL>", "<API Token>",
        WithHttpClient(&http.Client{Timeout: time.Minute}),
    )
    if err != nil {
        panic(err)
    }
    _ = client
}
```
