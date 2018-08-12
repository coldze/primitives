# Using service-info

Compile with `ldflags` like this:

```
go build -ldflags "-X github.com/coldze/primitives/service.version=<some-value> -X github.com/coldze/primitives/service.name=<some-name>"
```

Somewhere in your code:

```
version := service.GetVersion()
name := service.GetServiceName()
```