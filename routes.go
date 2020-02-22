package main

func (srv *server) routes() {
	srv.router.GET("/", srv.Index())
	srv.router.GET("/api/pincode/:pincode", srv.sendPincode())
	srv.router.GET("/api/pincode", srv.sendPincodeByCityAndDis())
	srv.router.GET("/healthz", srv.handleHealth())
}
