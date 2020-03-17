package server

type Model struct {
	Port int
}

func New(port int) *Model {
	server := &Model{
		Port: port,
	}
	return server
}
