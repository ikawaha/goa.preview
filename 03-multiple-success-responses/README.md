# Missing response transform: a method with more than one success response

## Symptom

A method whose HTTP mapping declares two non-error responses fails generation. Two shapes were
observed; both come from a missing entry in the planned response transforms.

**Shape A: two success responses, no `Tag`** (`variants/untagged.go`):

```
HTTP marshal conversion for update server response 202   default was planned but not rendered
```

**Shape B: two success responses with `Tag` and a named `Body`** (`variants/named_body.go`):

```
panic: runtime error: invalid memory address or nil pointer dereference
	goa.design/goa/v3/http/codegen.(*ServicesData).buildResponseResultInit  http/codegen/service_data.go:3216
	goa.design/goa/v3/http/codegen.(*ServicesData).buildResponses           http/codegen/service_data.go:3044
	goa.design/goa/v3/http/codegen.(*ServicesData).buildResultData          http/codegen/service_data.go:2894
	goa.design/goa/v3/http/codegen.(*ServicesData).analyze                  http/codegen/service_data.go:1263
```

## Steps to reproduce

```sh
go run goa.design/goa/v3/cmd/goa gen repro/design -o .   # shape A
./check.sh v3.31.0-preview.3                             # both shapes
```

| Design | preview.1 / preview.3 | v3.27.0 - v3.30.0 |
|:--|:--|:--|
| two success responses, no `Tag` | **failed** | OK |
| two success responses, `Tag` + `Body(attribute)` | **panic** | OK |
| two success responses, `Tag`, no `Body` | OK | OK |
| single success response | OK | OK |

## Diagnosis

`http/codegen/service_data.go:3215-3216`:

```go
transforms := sd.transforms.responses[viewedConstructorKey{endpoint: e, response: resp, view: bodyType.View}]
converted, helpers, err := sd.clientWireTypes.renderTransform(transforms.clientDecode, clientBody, "body", "v", transformctx, svcctx)
```

`responses` is `map[viewedConstructorKey]*plannedResponseTransforms` (`http/codegen/plan.go:718`),
so a missing key yields a nil pointer that is dereferenced on the next line. That is shape B.
Shape A is the same missing planning surfacing on the server side, where the unrendered conversion
is detected and reported instead.

Adding a nil guard and printing the key shows which combinations were never planned. On a real
service this reported both responses of the same method:

```
service=service-a  method=method-a  status=200  view="default"  tag=[]
service=service-a  method=method-a  status=202  view="default"  tag=[outcome Accepted]
service=service-b  method=method-b  status=200  view="default"  tag=[]
service=service-b  method=method-b  status=202  view="default"  tag=[outcome Accepted]
```

Both of those methods declare a named `Body` on each response, matching shape B.

A nil check alone would only hide the problem: the planning step needs to register a transform for
every (endpoint, response, view) combination the generator later renders.
