# cone 🗼

The goal of `cone` is to provide an easy to use event handler implementation
that can support a multitude of event backends.

---

### Table of contents

- [Usage](#usage)
- [Middleware](#middleware)
- [Todo](#todo)

---

# Usage

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

# Middleware

Middleware can be placed around a specific handler.

```go

middleware := func(next cone.Handler) cone.Handler {
    return func(r cone.Response, e *cone.Event) {
        next.Serve(r, e)
    }
}

var handler cone.HandlerFunc = func(r cone.Response, e *cone.Event) {
    _ = r.Ack()
}

h := cone.NewHandlerMux()
h.Handler("event.subject", middleware(handler))
```

Or the whole mux.

```go
middleware := func(next cone.Handler) cone.Handler {
    return func(r cone.Response, e *cone.Event) {
        next.Serve(r, e)
    }
}

s := conetest.NewSource()
h := cone.NewHandlerMux()
c := cone.New(s, middleware(h))
```

# Todo

- [ ] Event context
- [X] Event headers
- [X] Handler middleware
- [ ] Consumer subject wildcard `event.*`
- [ ] Source benchmark (Jetstream)
