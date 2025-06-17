package server

func (r *RouterHandler) ImageRouter() {

	imageRouter := r.Server.Group("/image")

	imageRouter.Static("/", "/app/images")
}
