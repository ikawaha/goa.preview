# Two untagged success responses: a planned conversion is never rendered

## Symptom

A method whose HTTP mapping declares two success responses without `Tag` stops generation:

```
HTTP marshal conversion for m server response 202   default was planned but not rendered
```

The result type does not matter: the same design with a plain `Type` fails the same way, so this is
not about views.

## Steps to reproduce

```sh
go run goa.design/goa/v3/cmd/goa gen repro/design -o .
./check.sh v3.31.0-preview.3   # condition matrix
./check.sh v3.30.0             # control
```

| Goa | Result |
|:--|:--|
| v3.31.0-preview.1 | failed |
| v3.31.0-preview.3 | failed |
| v3.30.0, v3.29.2, v3.28.0, v3.27.0 | generates successfully |

## Condition matrix

| Design | preview.1 / preview.3 | v3.27.0 - v3.30.0 |
|:--|:--|:--|
| two untagged success responses, result type | **failed** | OK |
| two untagged success responses, plain type | **failed** | OK |
| two success responses selected by `Tag` | OK | OK |
| single success response | OK | OK |

## Cause

Goa handles only the first untagged response. `http/codegen/plan.go:1798` skips the rest when it
declares the client result constructors:

```go
noTagSeen := false
for _, response := range endpoint.Responses {
	if response.Tag[0] == "" {
		if noTagSeen {
			continue
		}
		noTagSeen = true
	}
	...
}
```

The loop that collects the wire conversions, `http/codegen/service_data.go:1610`, has no such skip:

```go
for _, response := range endpoint.Responses {
	body := bodies.response(response)
	...
}
```

So a server conversion is planned for the second untagged response, nothing ever renders it, and
generation reports the unrendered conversion. Instrumenting both sides shows the asymmetry:

```
[COLLECT] method=m status=200 tag=[] view="default"
[COLLECT] method=m status=202 tag=[] view="default"   <- planned
[LOOKUP ] method=m status=200 tag=[] view="default" found=true
                                                      <- 202 is never looked up
HTTP marshal conversion for m server response 202   default was planned but not rendered
```

The keys match on both sides; the problem is that a conversion is planned for a response the
generator does not render.

## Note

A second, independent failure was originally reported together with this one: a response with an
explicit `Body(attribute)` and a result type panics with a nil pointer dereference. That one needs
only a single response and has a different cause (the collecting side registers an empty view for an
explicit body while the lookup uses the body view), so it is tracked separately.
