# cone 🗼

A generic event consumer with an `http.Server`-like interface.

Supports any event backend that implements the `Source` interface.

## Usage

```go
// Configure handlers
h := cone.NewHandlerMux()
h.HandleFunc("event.subject", func(r cone.Response, e *cone.Event) {
    _ = r.Ack()
})
h.HandleFunc("event.subject2", func(r cone.Response, e *cone.Event) {
    _ = r.Ack()
})

// Setup an event source
s := conetest.NewSource()

// Setup consumer with source and handler
c := cone.New(s, h)

// Start consumer. Processes events from the source and calls your handler.
c.ListenAndConsume()
```

## Middleware

Middleware can be placed around a specific handler.

```go

middleware := func(next cone.Handler) cone.Handler {
    return func(r cone.Response, e *cone.Event) {
        next.Serve(r, e)
    }
}


h := cone.NewHandlerMux()
h.HandlerFunc("event.subject", middleware())
```

Or the whole mux.

```go
globalMiddleware := func(next cone.Handler) cone.Handler {
    return func(r cone.Response, e *cone.Event) {
        next.Serve(r, e)
    }
}

h := cone.NewHandlerMux()
s := conetest.NewSource()
c := cone.New(s, globalMiddleware(h))
```

# Todo

- [ ] Event context
- [ ] Event headers
- [X] Handler middleware
- [ ] Consumer subject wildcard `event.*`
- [ ] Source benchmark (Jetstream)
