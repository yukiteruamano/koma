# mangodex

Golang API wrapper for MangaDex v5's MVP API.

Full API documentation is found [here](https://api.mangadex.org).

> Fork mantenido de [`darylhjd/mangodex`](https://github.com/darylhjd/mangodex) — proyecto original abandonado.

This branch contains only essential services, such as Manga searching and image downloading.

## Installation
To install, do `go get -u github.com/yukiteruamano/mangodex@essential`.

## Usage
```golang
package main

import (
	"fmt"
	
	m "github.com/yukiteruamano/mangodex@simple"
)

func main() {
	// Create new client.
	// Without logging in, you may not be able to access 
	// all API functionality.
	c := m.NewDexClient()

	// Login using your username and password.
	err := c.Auth.Login("user", "password")
	if err != nil {
		fmt.Println("Could not login!")
	}
}
```

## Contributing
Any contributions are welcome.
