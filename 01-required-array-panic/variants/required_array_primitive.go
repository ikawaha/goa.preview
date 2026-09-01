package design

import . "goa.design/goa/v3/dsl"

var Res = ResultType("application/vnd.repro+json", func() {
	Attributes(func() {
		Attribute("items", ArrayOf(String))
		Required("items")
	})
})

var _ = Service("svc", func() {
	Method("m", func() {
		Result(Res)
		HTTP(func() { GET("/") })
	})
})
