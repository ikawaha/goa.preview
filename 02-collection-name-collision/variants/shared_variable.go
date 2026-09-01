package design

import . "goa.design/goa/v3/dsl"

var RT = ResultType("application/vnd.reservation+json", func() {
	Attributes(func() {
		Attribute("id", Int)
		Required("id")
	})
})

// CollectionOf is called once and the value is shared: this works.
var RTC = CollectionOf(RT)

var _ = Service("reservation", func() {
	Method("search1", func() {
		Result(RTC)
		HTTP(func() { GET("/a") })
	})
	Method("search2", func() {
		Result(RTC)
		HTTP(func() { GET("/b") })
	})
})
