package design

import . "goa.design/goa/v3/dsl"

var RT = Type("Res", func() {
	Attribute("name", String)
	Required("name")
})

var _ = Service("svc", func() {
	Method("m", func() {
		Result(RT)
		HTTP(func() {
			POST("/")
			Response(StatusOK)
			Response(StatusAccepted)
		})
	})
})
