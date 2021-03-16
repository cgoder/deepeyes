package gss

func NewService() *ServerST {
	svr := new(ServerST)
	svr.LoadConfig()
	svr.LoadDB()
	return svr
}
