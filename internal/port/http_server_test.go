package port_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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
	_ = os.Setenv("JWT_PRIVATE_KEY", "testfiles/certs/privateKey.der")
}

type HTTPServerSuite struct {
	suite.Suite
	mux *chi.Mux
}

func (s *HTTPServerSuite) SetupTest() {

	a := assert.New(s.T())

	_ = os.Mkdir("testfiles", os.ModePerm)

	_ = os.Mkdir("testfiles/certs", os.ModePerm)

	_ = os.Mkdir("testfiles/db", os.ModePerm)

	f, e := os.Create("testfiles/db/chat.db")
	a.NoError(e)
	_ = f.Close()

	pk, e := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	a.NoError(e)

	pb, e := x509.MarshalECPrivateKey(pk)
	a.NoError(e)

	bb := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pb})
	a.NoError(e)

	e = os.WriteFile("testfiles/certs/privateKey.der", bb, 0600)
	a.NoError(e)

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

func (s *HTTPServerSuite) TestUserLogin() {

	a := assert.New(s.T())

	bb, e := json.Marshal(&port.NewUserParams{
		NewUserPayload: &port.NewUserPayload{
			Username: "john.doe",
			Email:    "john.doe@mailbox.com",
			Password: "test123",
		},
	})
	a.NoError(e)

	rq, e := http.NewRequest(http.MethodPost, "/api/v1/user", bytes.NewReader(bb))
	a.NoError(e)

	rq.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, rq)

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

func TestHTTPServerSuite(t *testing.T) {
	suite.Run(t, new(HTTPServerSuite))
}
