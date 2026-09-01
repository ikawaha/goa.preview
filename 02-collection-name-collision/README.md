# Name collision: CollectionOf called twice in one service

## Symptom

Calling `CollectionOf` twice for the same result type inside one service fails generation:

```
declare user type "ReservationCollection": generated package "repro/gen/reservation"
cannot declare exact type "ReservationCollection": already declared by exact type
```

## Steps to reproduce

```sh
go run goa.design/goa/v3/cmd/goa gen repro/design -o .
```

| goa | Result |
|:--|:--|
| v3.31.0-preview.1 | failed |
| v3.31.0-preview.3 | failed |
| v3.30.0 | generates successfully |
| v3.29.2 | generates successfully |
| v3.28.0 | generates successfully |
| v3.27.0 | generates successfully |

`./check.sh <version>` runs the matrix below against one version.

## Condition matrix

| Design | preview.1 / preview.3 | v3.27.0 - v3.30.0 |
|:--|:--|:--|
| two methods in one service, `CollectionOf` called twice | **failed** | OK |
| same, but the `CollectionOf` value is stored in a variable and shared | OK | OK |
| two services, `CollectionOf` called once in each | OK | OK |
| identifier without a `+json` suffix, `CollectionOf` called twice | OK | OK |

The last row is the interesting one: the only difference is the media type suffix.

## Cause

`CollectionOf` is supposed to reuse an existing collection type, but the lookup never matches when
the identifier has a suffix such as `+json`.

```go
// dsl/result_type.go
canonical := expr.CanonicalIdentifier(id)   // "application/vnd.reservation; type=collection"
if mt := expr.GeneratedResultType(canonical); mt != nil {
	// Already have a type for this collection, reuse it.
	return mt
}

// expr/result_types_root.go
func GeneratedResultType(id string) *ResultTypeExpr {
	for _, rt := range *GeneratedResultTypes {
		if rt.Identifier == id {   // stored raw: "application/vnd.reservation+json; type=collection"
			return rt
		}
	}
	return nil
}
```

`CanonicalIdentifier` strips the `+json` suffix, but the stored identifiers keep it, so a
canonicalized key is compared against raw values and never matches. Every `CollectionOf` call
therefore creates a new result type. Instrumented output:

```
lookup key    : "application/vnd.reservation; type=collection"
stored entry  : "application/vnd.reservation+json; type=collection"  (canonical: "application/vnd.reservation; type=collection")
-> no match, creating a new type
```

The two resulting expressions are distinct values that both want the Go name
`ReservationCollection`:

```
new     : ResultType ptr=0x...31d0 name="ReservationCollection" identifier="application/vnd.reservation+json; type=collection"
existing: ResultType ptr=0x...3130 name="ReservationCollection" identifier="application/vnd.reservation+json; type=collection"
```

`DeclareUserType` (`codegen/generated_types.go`) deduplicates by `userType.Origin()`, so the two
distinct expressions are not merged and the second `DeclareName` collides.

This code is identical in v3.30, where the duplicate was absorbed later during generation. The
v3.31 planner assigns names up front and reports the collision instead.

This also does not match the documented behavior for collisions. `codegen/ARCHITECTURE.md` states
that "real name collisions receive stable numeric suffixes instead of suffixes determined by
discovery order", and describes failing generation only for unions: "Goa never adds a numeric suffix
to resolve a collision. Separately authored unions that ask for the same family name now fail
generation". This design has no union and declares no duplicate; the duplicate is created inside
`CollectionOf`.


## Suggested fix (one line)

Compare canonical identifiers on both sides. There are two callers: `dsl/result_type.go` passes a
canonicalized value, `expr/dup.go:152` passes a raw one, so canonicalizing inside the lookup keeps
both correct.

```diff
 func GeneratedResultType(id string) *ResultTypeExpr {
+	canonical := CanonicalIdentifier(id)
 	for _, rt := range *GeneratedResultTypes {
-		if rt.Identifier == id {
+		if CanonicalIdentifier(rt.Identifier) == canonical {
 			return rt
 		}
 	}
 	return nil
 }
```

Verified with this change:

- all designs in `variants/` generate
- the Goa test suite shows no new failures
- two services of a real multi-service project that previously failed here now generate

## Workaround without patching Goa

This is only a way for testers to keep moving on the preview. Calling `CollectionOf` twice is valid
and generates correctly on every released version, so the design itself does not need to change.

Call `CollectionOf` once and share the value:

```go
var RTC = CollectionOf(RT)

Method("search1", func() { Result(RTC) /* ... */ })
Method("search2", func() { Result(RTC) /* ... */ })
```
