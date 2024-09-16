# cone 🗼

# Usage

```go
s := conetest.NewSource()
c := cone.New(s)
c.HandleFunc("event.subject", func(r cone.Response, e cone.Event) {
    _ = r.Ack()
})

c.ListenAndConsume()
```

# Todo

- [ ] Event context
- [ ] Event headers
- [ ] Handler middleware
- [ ] Consumer subject wildcard `event.*`
- [ ] Source benchmark (Jetstream)
