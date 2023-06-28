package port_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/trevatk/go-chat/internal/domain"
	"github.com/trevatk/go-chat/internal/port"
	"github.com/trevatk/go-chat/internal/port/middleware"
	"github.com/trevatk/go-pkg/db"
)

func init() {
	_ = os.Setenv("SQLITE_DSN", "testfiles/db/chat.db")
	_ = os.Setenv("SQLITE_MIGRATIONS_DIR", "../../migrations")
	_ = os.Setenv("JWT_PRIVATE_KEY", "testfiles/certs/ec-secp384r1-priv-key.der")
}

type HTTPServerSuite struct {
	suite.Suite
	mux *chi.Mux
}

func (s *HTTPServerSuite) SetupTest() {

	a := assert.New(s.T())

	sdb, e := db.NewSQLite()
	a.NoError(e)

	e = db.MigrateSQLite(sdb)
	a.NoError(e)

	b := domain.NewBundle(sdb)

	srv, e := port.NewHTTPServer(b)
	a.NoError(e)

	mw, e := middleware.NewAuthenticator()
	a.NoError(e)

	s.mux = port.NewRouter(srv, mw)
}

func (s *HTTPServerSuite) TestCreateUser() {

	a := assert.New(s.T())

	cases := []struct {
		expected int
		payload  *port.NewUserParams
	}{
		{
			// success
			expected: http.StatusCreated,
			payload: &port.NewUserParams{
				NewUserPayload: &port.NewUserPayload{
					Username: "john.doe",
					Email:    "john.doe@mailbox.com",
					Password: "test123",
				},
			},
		},
	}

	for _, c := range cases {

		bb, e := json.Marshal(c.payload)
		a.NoError(e)

		rq, e := http.NewRequest(http.MethodPost, "/api/v1/user", bytes.NewReader(bb))
		a.NoError(e)

		rq.Header.Add("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		s.mux.ServeHTTP(rr, rq)

		a.Equal(c.expected, rr.Code)
	}
}

func (s *HTTPServerSuite) TestUserLogin() {

	a := assert.New(s.T())

	cases := []struct {
		expected int
		payload  *port.UserLoginRequest
	}{
		{
			// success username
			expected: http.StatusAccepted,
			payload: &port.UserLoginRequest{
				Username: "john.doe",
				Password: "test123",
			},
		},
		{
			// success email
			expected: http.StatusAccepted,
			payload: &port.UserLoginRequest{
				Username: "john.doe@mailbox.com",
				Password: "test123",
			},
		},
	}

	for _, c := range cases {

		bb, e := json.Marshal(c.payload)
		a.NoError(e)

		rq, e := http.NewRequest(http.MethodPost, "/api/v1/user/login", bytes.NewReader(bb))
		a.NoError(e)

		rq.Header.Add("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		s.mux.ServeHTTP(rr, rq)

		a.Equal(c.expected, rr.Code)
	}
}

func TestHTTPServerSuite(t *testing.T) {
	suite.Run(t, new(HTTPServerSuite))
}
