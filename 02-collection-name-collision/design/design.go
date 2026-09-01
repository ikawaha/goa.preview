package design

import . "goa.design/goa/v3/dsl"

var RT = ResultType("application/vnd.reservation+json", func() {
	Attributes(func() {
		Attribute("id", Int)
		Required("id")
	})
})

var _ = Service("reservation", func() {
	Method("search1", func() {
		Result(CollectionOf(RT))
		HTTP(func() {
			GET("/a")
		})
	})
	Method("search2", func() {
		Result(CollectionOf(RT))
		HTTP(func() {
			GET("/b")
		})
	})
})