package session

import (
	"net/http"
	"sync"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/honmaple/forest"
)

const sessionName = "_session"

type session struct {
	name     string
	store    sessions.Store
	session  *sessions.Session
	request  *http.Request
	response http.ResponseWriter
	written  bool
}

func (s *session) Get(key interface{}) interface{} {
	return s.Session().Values[key]
}

func (s *session) Set(key interface{}, val interface{}) {
	s.Session().Values[key] = val
	s.written = true
}

func (s *session) Delete(key interface{}) {
	delete(s.Session().Values, key)
	s.written = true
}

func (s *session) Clear() {
	sess := s.Session()
	for key := range sess.Values {
		delete(sess.Values, key)
	}
	s.written = true
}

func (s *session) AddFlash(value interface{}, vars ...string) {
	s.Session().AddFlash(value, vars...)
	s.written = true
}

func (s *session) Flashes(vars ...string) []interface{} {
	s.written = true
	return s.Session().Flashes(vars...)
}

func (s *session) Save() error {
	if s.Written() {
		e := s.Session().Save(s.request, s.response)
		if e == nil {
			s.written = false
		}
		return e
	}
	return nil
}

func (s *session) Session() *sessions.Session {
	if s.session == nil {
		sess, err := s.store.Get(s.request, s.name)
		if err != nil {
			return nil
		}
		s.session = sess
	}
	if s.session.Options == nil {
		s.session.Options = &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 7,
		}
	}
	return s.session
}

func (s *session) Written() bool {
	return s.written
}

func (s *session) reset(r *http.Request, w http.ResponseWriter) {
	s.request = r
	s.response = w
	s.session = nil
	s.written = false
}

func Session(c forest.Context) *session {
	s := c.Get(sessionName)
	if s == nil {
		return nil
	}
	return s.(*session)
}

func Middleware(name string, store sessions.Store) forest.HandlerFunc {
	sessionPool := sync.Pool{
		New: func() interface{} {
			return &session{name: name, store: store}
		},
	}
	return func(c forest.Context) error {
		s := sessionPool.Get().(*session)
		s.reset(c.Request(), c.Response())
		defer sessionPool.Put(s)

		c.Set(sessionName, s)
		defer context.Clear(c.Request())
		return c.Next()
	}
}
