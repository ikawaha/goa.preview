package design

import . "goa.design/goa/v3/dsl"

var Item = Type("Item", func() {
	Attribute("id", Int)
	Required("id")
})

var RT = ResultType("application/vnd.repro+json", func() {
	Attributes(func() {
		Attribute("outcome", String)
		Attribute("items", ArrayOf(Item))
		Required("outcome", "items")
	})
})

var _ = Service("svc", func() {
	Method("overwrite", func() {
		Result(RT)
		HTTP(func() {
			POST("/")
			Response(StatusOK, func() {
				Body("items")
			})
			Response(StatusAccepted, func() {
				Tag("outcome", "Accepted")
				Body("items")
			})
		})
	})
})
