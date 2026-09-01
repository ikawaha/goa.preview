# Panic: required array of a user type inside a result type

## Symptom

`goa gen` panics when a ResultType has a **required attribute of type `ArrayOf(<user type>)`**.

```
panic: resolve HTTP type "Inner": plan Go named type "Inner": HTTP type "Inner" has no planned declaration
	goa.design/goa/v3/http/codegen.(*wireAttributeScope).Ref              http/codegen/wire_catalog.go:1706
	goa.design/goa/v3/codegen.transformObject.func2                       codegen/go_transform.go:1153
	goa.design/goa/v3/http/codegen.(*ServicesData).buildResponseBodyType  http/codegen/service_data.go:3796
```

## Steps to reproduce

```sh
go run goa.design/goa/v3/cmd/goa gen repro/design -o .
```

| goa | Result |
|:--|:--|
| v3.31.0-preview.1 | panic |
| v3.31.0-preview.3 | panic |
| v3.30.0 | generates successfully |
| v3.29.2 | generates successfully |
| v3.28.0 | generates successfully |
| v3.27.0 | generates successfully |

`./check.sh <version>` runs the whole matrix below against one version.

## Condition matrix

Verified with `./check.sh v3.31.0-preview.3` (designs in `variants/`):

| Design | preview.1 / preview.3 | v3.27.0 - v3.30.0 |
|:--|:--|:--|
| ResultType + Required + `ArrayOf(user type)` | **panic** | OK |
| ResultType + optional + `ArrayOf(user type)` | OK | OK |
| ResultType + Required + `ArrayOf(String)` | OK | OK |
| ResultType + Required + user type (not an array) | OK | OK |
| plain Type (not a ResultType) + Required + `ArrayOf(user type)` | OK | OK |
| ResultType + Required + `ArrayOf(ArrayOf(user type))` | **panic** | OK |
| ResultType with explicit views + Required + `ArrayOf(user type)` | **panic** | OK |

No code generation plugins are involved: this reproduces with plain Goa.

## Cause

`codegen/go_transform.go:1153`. When a required array field is nil, the generator emits
`<target> = []Elem{}`. To spell the element type it calls `Ref` on the **outer** context `ta`
instead of `fieldAttrs`, the field-level context created a few lines above at 1074.

```go
fieldAttrs := enterTransformAttrs(srcc, tgtc, ta)   // line 1074
...
if expr.IsArray(srcc.Type) && srcMatt.IsRequired(n) {
    code += fmt.Sprintf("else {\n\t%s = []%s{}\n}\n", tgtVar,
        ta.TargetCtx.Scope.Ref(...))                // line 1153: uses ta
}
```

This call site is unchanged since v3.30 (`go_transform.go:375` there), where it was harmless
because `Scope.Ref` only spelled a name. In v3.31 `Ref` resolves a declaration planned ahead of
generation and requires an exact identity match. Because this call does not go through `Enter()`,
`policy.view` is still `"default"`, so no planned record matches and the lookup panics.
The collecting side (`http/codegen/wire_catalog.go`, `collectRecursive`) always registers nested
types with `policy.view = ""`.

Observed identity mismatch:

```
looked up: preferred="Inner" sourceID=""      role=3 policy.view="default"
planned:   preferred="Inner" sourceID="Inner" role=0 policy.view=""
```

This also explains the narrow trigger: optional arrays never reach the branch, primitive elements
never resolve a planned declaration, and a plain Type carries no view.

## Suggested fix (one line)

```diff
-        ta.TargetCtx.Scope.Ref(expr.AsArray(tgtc.Type).ElemType, ta.TargetCtx.Pkg(expr.AsArray(tgtc.Type).ElemType)))
+        fieldAttrs.TargetCtx.Scope.Ref(elem, fieldAttrs.TargetCtx.Pkg(elem)))
```

Every other call in the same block (lines 1086, 1094-1108, 1171, 1187) already uses `fieldAttrs`;
line 1153 looks like the one that was missed.

With this change all three panicking designs above generate, the generated code compiles, and the
emitted expression is correct (`body.Items = []*InnerResponseBody{}`). A real multi-service design
(11 services, 87 generated files) also generates and builds.

## Other call sites

`go_transform.go` has other calls that use the outer `ta`, but none of them fail here:
lines 1295 (`transformArray`) and 1332-1333 (`transformMap`) receive an already-entered context from
their caller, and lines 997 and 1205 are only used for primitive aliases, which are not view
sensitive. This was checked by making `Ref` report instead of panic and running the whole Goa test
suite plus the designs in `variants/`: line 1153 was the only reported call site.
