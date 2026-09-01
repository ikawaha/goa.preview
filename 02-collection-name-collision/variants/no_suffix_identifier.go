package design

import . "goa.design/goa/v3/dsl"

// The identifier has no "+json" suffix, so CanonicalIdentifier() returns it
// unchanged and the CollectionOf cache lookup happens to match.
var RT = ResultType("application/vnd.reservation", func() {
	Attributes(func() {
		Attribute("id", Int)
		Required("id")
	})
})

var _ = Service("reservation", func() {
	Method("search1", func() {
		Result(CollectionOf(RT))
		HTTP(func() { GET("/a") })
	})
	Method("search2", func() {
		Result(CollectionOf(RT))
		HTTP(func() { GET("/b") })
	})
})
