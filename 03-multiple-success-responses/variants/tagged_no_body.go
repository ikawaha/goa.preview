package design

import . "goa.design/goa/v3/dsl"

var RT = ResultType("application/vnd.repro+json", func() {
	Attributes(func() {
		Attribute("outcome", String)
		Attribute("id", Int)
		Required("outcome", "id")
	})
})

var _ = Service("svc", func() {
	Method("update", func() {
		Result(RT)
		HTTP(func() {
			POST("/")
			Response(StatusOK)
			Response(StatusAccepted, func() {
				Tag("outcome", "Accepted")
			})
		})
	})
})
