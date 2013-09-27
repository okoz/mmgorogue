package main

type ChatService interface {
	Send(origin Entity, message string)
	Register(e Entity, handler func(Entity, string))
}

type chatService struct {
	handlers map[Entity]func(Entity, string)
}

func CreateChatService() ChatService {
	s := &chatService{make(map[Entity]func(Entity, string))}
	return s
}

func (s *chatService) Send(origin Entity, message string) {
	for e, handler := range s.handlers {
		if e != origin {
			handler(origin, message)
		}
	}
}

func (s *chatService) Register(e Entity, handler func(Entity, string)) {
	s.handlers[e] = handler
}

func (s *chatService) Unregister(e Entity) {
	delete(s.handlers, e)
}