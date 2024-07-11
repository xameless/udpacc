package inbound

type ListenConfig interface {
	Config()
}

type Inbound interface {
	Listen(ListenConfig)
}
