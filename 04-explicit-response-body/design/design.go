package design

import . "goa.design/goa/v3/dsl"

var RT = ResultType("application/vnd.repro+json", func() {
	Attributes(func() {
		Attribute("name", String)
		Required("name")
	})
})

var _ = Service("svc", func() {
	Method("m", func() {
		Result(RT)
		HTTP(func() {
			POST("/")
			Response(StatusOK, func() {
				Body("name")
			})
		})
	})
})