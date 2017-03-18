golove
======

This repository contains two Go packages. The first, `love`, is a client library
for [Yelp Love](https://github.com/Yelp/love). The second, `golove`, is a
program that allows you to send love from the command line.

Documentation is available at [godoc.org](https://godoc.org):
- [`love`](https://godoc.org/github.com/hacsoc/golove/love)
- [`golove`](https://godoc.org/github.com/hacsoc/golove/golove)

To use either tool, you must have an API token. API tokens are available only to
administrators, since they allow you to send love as any user. To create an API
token, go to your website. Select "API Keys" from the Admin dropdown, type a
description, and hit "Add". Then copy the generated key.

Usage
-----

The library is fairly simple, and `golove` is a good reference for library
usage. Here is a simplified example:

```go
import "github.com/hacsoc/golove/love"

// ...

client := love.Client(api_key, base_url)

// send love from hammy to darwin
err := client.SendLove("hammy", "darwin", "great job fixing the site!")
if err != nil {
	// handle error
}

// returns loves sent to darwin, limit of 20
loves, err := love.GetLove("", "darwin", 20)
if err != nil {
	// handle error
}

// gets completions as the user types "ha"
users, err := love.Autocomplete("ha")
if err != nil {
	// handle error
}
```

Possible causes for errors are network connectivity issues, or HTTP status codes
which correspond to failures.
