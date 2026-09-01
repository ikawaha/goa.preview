package design

import . "goa.design/goa/v3/dsl"

var RT = ResultType("application/vnd.reservation+json", func() {
	Attributes(func() {
		Attribute("id", Int)
		Required("id")
	})
})

// Two calls, but in different services, so different generated packages.
var _ = Service("svc1", func() {
	Method("search", func() {
		Result(CollectionOf(RT))
		HTTP(func() { GET("/a") })
	})
})

var _ = Service("svc2", func() {
	Method("search", func() {
		Result(CollectionOf(RT))
		HTTP(func() { GET("/b") })
	})
})
