package design

import . "goa.design/goa/v3/dsl"

var Item = Type("Item", func() {
	Attribute("id", Int)
	Required("id")
})

var RT = ResultType("application/vnd.repro+json", func() {
	Attributes(func() {
		Attribute("item", Item)
		Required("item")
	})
})

var _ = Service("svc", func() {
	Method("m", func() {
		Result(RT)
		HTTP(func() {
			POST("/")
			Response(StatusOK, func() {
				Body("item")
			})
		})
	})
})
