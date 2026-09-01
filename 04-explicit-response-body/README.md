# Panic: a response body set to one result attribute

## Symptom

`goa gen` panics when a response sets its body to one attribute of a result type:

```
panic: runtime error: invalid memory address or nil pointer dereference
	goa.design/goa/v3/http/codegen.(*ServicesData).buildResponseResultInit  http/codegen/service_data.go:3216
	goa.design/goa/v3/http/codegen.(*ServicesData).buildResponses           http/codegen/service_data.go:3044
	goa.design/goa/v3/http/codegen.(*ServicesData).buildResultData          http/codegen/service_data.go:2894
	goa.design/goa/v3/http/codegen.(*ServicesData).analyze                  http/codegen/service_data.go:1263
	goa.design/goa/v3/http/codegen.(*Plan).link                             http/codegen/plan.go:1968
```

One response is enough. The same design with a plain `Type` result generates.

## Steps to reproduce

```sh
go run goa.design/goa/v3/cmd/goa gen repro/design -o .
./check.sh v3.31.0-preview.3   # condition matrix
./check.sh v3.30.0             # control
```

`design/design.go`:

```go
package design

import . "goa.design/goa/v3/dsl"

var RT = ResultType("application/vnd.repro+json", func() {
	Attributes(func() {
		Attribute("name", String)
		Required("name")
	})
})

var _ = Service("svc", func() {
	Method("m", func() {
		Result(RT)
		HTTP(func() {
			POST("/")
			Response(StatusOK, func() {
				Body("name")
			})
		})
	})
})
```

| Goa | Result |
|:--|:--|
| v3.31.0-preview.1 | panic |
| v3.31.0-preview.3 | panic |
| v3.30.0, v3.29.2, v3.28.0, v3.27.0 | generates successfully |

## Condition matrix

| Design | preview.1 / preview.3 | v3.27.0 - v3.30.0 |
|:--|:--|:--|
| response `Body(attribute)`, result type | **panic** | OK |
| response `Body(attribute)` holding a user type, result type | **panic** | OK |
| response `Body(attribute)`, result type with explicit views | OK | OK |
| response `Body(attribute)`, plain type | OK | OK |
| no response `Body`, result type | OK | OK |

## Cause

Two switches decide the same thing and order their cases differently.

The collecting side, `http/codegen/service_data.go` in `collectPlannedTransforms`, settles an
explicit body attribute before a selected view:

```go
selected := clientResponseViewNameExpr(endpoint, resultType)
switch {
case origin != "":            // Body("name")
	empty := ""
	clientViews = append(clientViews, &empty)
case selected != "":
	clientViews = append(clientViews, &selected)
case explicitBody:            // Body(func() { ... })
	empty := ""
	clientViews = append(clientViews, &empty)
default:
	...
}
```

The consuming side, `buildResponses`, checks the selected view first:

```go
switch {
case clientView != "":
	...
	clientBodyView = &clientView
case origin != "" || explicitBody:
	...
	clientBodyView = &vname
default:
	...
}
```

`Body("name")` sets both `origin` and a selected client view, so the two sides disagree: collection
records the transform under an empty view while the lookup at `service_data.go:3215` asks for
`"default"`. `responses` is `map[viewedConstructorKey]*plannedResponseTransforms`
(`http/codegen/plan.go:718`), so the miss yields a nil pointer that is dereferenced on the next
line. Instrumenting both sides:

```
[COLLECT] method=m status=200 tag=[] view=""
[LOOKUP ] method=m status=200 tag=[] view="default" found=false
```

The variants explain the rest of the matrix: with explicit views the selected view is empty, so the
order does not matter; a plain type has no views at all; and without an explicit body `origin` is
empty, so both sides take the same branch.

## Suggested fix

Move only the `origin` case ahead of the selected view on the consuming side, so both switches
settle an explicit body attribute first:

```diff
 					switch {
+					case origin != "":
+						clientBodyData = sds.buildResponseBodyType(respBody, result, e, false, &vname, sd, nil, resultOwner, bodyOwner)
+						clientBodyView = &vname
 					case clientView != "":
 						clientRespBody = effectiveClientResponseBodyForView(respBody, clientView, e)
 						clientBodyData = sds.buildResponseBodyType(clientRespBody, result, e, false, &clientView, sd, nil, resultOwner, bodyOwner)
 						clientBodyView = &clientView
-					case origin != "" || explicitBody:
+					case explicitBody:
 						clientBodyData = sds.buildResponseBodyType(respBody, result, e, false, &vname, sd, nil, resultOwner, bodyOwner)
 						clientBodyView = &vname
```

Moving `explicitBody` as well breaks `TestBodyTypeInit/result-explicit-body-object`: an inline
`Body(func() { ... })` sets `explicitBody` without `origin`, and collection settles that case after
the selected view. Only the `origin` case belongs in front.

With this change all five designs above generate, the generated client decoder is the same as the
one produced by v3.30.0, and the Goa test suite shows no new failures.
