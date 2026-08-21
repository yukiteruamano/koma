# Gache

Gache is a dead simple file-based (or in-memory) cache library for Go with zero dependencies.

There are great caching libraries out there, but none of them fit my needs or are as simple as I'd like.
So I decided to write my own and share it with the world. 🐳

> **Fork note:** This fork moves the original module `github.com/metafates/gache` to `github.com/yukiteruamano/gache` and bumps the minimum Go version to **Go 1.26**. To migrate, replace the import:
> `github.com/metafates/gache` → `github.com/yukiteruamano/gache`.

## Installation

```bash
go get github.com/yukiteruamano/gache
```

Requires Go 1.26+.

## Usage Example

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/yukiteruamano/gache"
	"net/http"
	"time"
)

type Pokemon struct {
	Height int
}

// Create new cache instance
var cache = gache.New[map[string]*Pokemon](&gache.Options{
	// Path to cache file
	// If not set, cache will be in-memory
	Path: ".cache/pokemons.json",

	// Lifetime of cache.
	// If not set, cache will never expire
	Lifetime: time.Hour,
})

// getPokemon will get a pokemon by name from API
// Gonna Cache Em' All!
func getPokemon(name string) (*Pokemon, error) {
	// check if Pokémon is in cache
	pokemons, expired, err := cache.Get()
	if err != nil {
		return nil, err
	}
	
	if pokemons == nil {
	    pokemons = make(map[string]*Pokemon)
	}

	// if cache is expired, or Pokémon wasn't cached
	// Fetch it from API
	if pokemon, ok := pokemons[name]; !expired && ok {
		return pokemon, nil
	}

	// bla-bla-bla, boring stuff, etc...
	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + name)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var pokemon Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&pokemon); err != nil {
		return nil, err
	}

	// okay, we got our Pokémon, let's cache it
	pokemons[name] = &pokemon
	_ = cache.Set(pokemons)

	return &pokemon, nil
}

func main() {
	start := time.Now()
	for i := 0; i < 3; i++ {
		_, _ = getPokemon("pikachu")
	}
	fmt.Println(time.Since(start))
}
```

## New APIs (v1 hardening)

```go
c := gache.New[string](&gache.Options{Path: "cache.json", Lifetime: time.Hour})

// Clear cache
_ = c.Clear()

// Check expiration without mutating
if c.IsExpired() { /* ... */ }

// Expiration hook (called outside the lock to avoid deadlock)
c2 := gache.New[string](&gache.Options{
    Lifetime: time.Minute,
    ExpirationHook: func() { fmt.Println("expired") },
})

// Alternative encoder (gob)
opts := &gache.Options{Path: "cache.gob"}
gache.WithGob(opts)
c3 := gache.New[MyStruct](opts)

// Or implement a custom Encoder/Decoder
type MyCodec struct{}
func (MyCodec) Encode(w io.Writer, v any) error { /* ... */ return nil }
func (MyCodec) Decode(r io.Reader, v any) error { /* ... */ return nil }
```

## Security and Robustness

- Atomic writes (`*.tmp` + `Rename`) to avoid corruption on crash.
- Permissions `0755` / `0644` instead of `0777` / `0666`.
- Proper FD closing and handling of empty/malformed files.
- `Get()` now clears expired state and fires `ExpirationHook` outside the lock (do not call `Get`/`Set` synchronously inside the hook).

## Development

```bash
make test   # go test -race
make vet
make lint   # golangci-lint
make cover  # coverage
```
