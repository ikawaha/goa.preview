package design

import . "goa.design/goa/v3/dsl"

var RT = ResultType("application/vnd.repro+json", func() {
	Attributes(func() {
		Attribute("id", Int)
		Required("id")
	})
})

var _ = Service("svc", func() {
	Method("update", func() {
		Result(RT)
		HTTP(func() {
			POST("/")
			Response(StatusOK)
		})
	})
})
