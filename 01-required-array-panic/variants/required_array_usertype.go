package design

import . "goa.design/goa/v3/dsl"

var Inner = Type("Inner", func() {
	Attribute("a", String)
})

var Res = ResultType("application/vnd.repro+json", func() {
	Attributes(func() {
		Attribute("items", ArrayOf(Inner))
		Required("items")
	})
})

var _ = Service("svc", func() {
	Method("m", func() {
		Result(Res)
		HTTP(func() { GET("/") })
	})
})
