// Package port exposed endpoints
package port

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	"github.com/trevatk/go-chat/internal/domain"
	mw "github.com/trevatk/go-chat/internal/port/middleware"
	"github.com/trevatk/go-pkg/logging"
)

// HTTPServer exposed endpoints
type HTTPServer struct {
	bundle     *domain.Bundle
	privateKey *ecdsa.PrivateKey
}

// NewHTTPServer create new http server instance
func NewHTTPServer(bundle *domain.Bundle) (*HTTPServer, error) {

	p := os.Getenv("JWT_PRIVATE_KEY")
	if p == "" {
		return nil, errors.New("$JWT_PRIVATE_KEY is not set")
	}

	bb, e := os.ReadFile(filepath.Clean(p))
	if e != nil {
		return nil, fmt.Errorf("failed to open ecdsa private key %v", e)
	}

	b, _ := pem.Decode(bb)

	pk, e := x509.ParseECPrivateKey(b.Bytes)
	if e != nil {
		return nil, fmt.Errorf("failed to to parse key from bytes %v", e)
	}

	return &HTTPServer{bundle: bundle, privateKey: pk}, nil
}

// NewRouter create new chi implementation of http.ServeMux
func NewRouter(srv *HTTPServer, auth *mw.Authenticator) *chi.Mux {

	ao := os.Getenv("ALLOWED_ORIGINS")
	if ao == "" {
		ao = "http://localhost"
	}

	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{ao},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/api/v1", func(r chi.Router) {

		r.Use(auth.ValidateJWT)

		r.Get("/user/{user_id}", srv.fetchUser)
		r.Put("/user", srv.updateUser)
		r.Get("/user/search/{search_str}", srv.searchUsers)

		r.Post("/contact", srv.addContact)
		r.Get("/contact/search/{search_str}", srv.searchContacts)
		r.Get("/contact/{contact_id}", srv.fetchContact)
		r.Get("/contact/", srv.listContacts)
		r.Delete("/contact/{contact_id}", srv.deleteContact)

		r.Post("/conversation", srv.createConversation)
		r.Get("/conversation/", srv.listConversations)
	})

	r.Post("/api/v1/user", srv.createUser)
	r.Post("/api/v1/user/login", srv.userLogin)
	r.Get("/health", srv.health)

	return r
}

// NewUserPayload http new user payload model
type NewUserPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewUserParams new user request parameters
type NewUserParams struct {
	*NewUserPayload `json:"new_user"`
}

// Bind parse request into new user model
func (nup *NewUserParams) Bind(_ *http.Request) error {

	if nup.NewUserPayload == nil {
		return errors.New("missing new user parameters")
	}

	if nup.Username == "" {
		return errors.New("no name parameter provided")
	} else if nup.Email == "" {
		return errors.New("no email parameter provided")
	} else if nup.Password == "" {
		return errors.New("no password parameter provided")
	}

	return nil
}

// NewUserResponse new user response model
type NewUserResponse struct {
	*UserPayload `json:"user"`
}

// UserPayload http user model
type UserPayload struct {
	UID       string    `json:"uid"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserDetailsPayload http user details model
type UserDetailsPayload struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
}

func newUserResponse(user *domain.User) *NewUserResponse {
	return &NewUserResponse{
		UserPayload: &UserPayload{
			UID:       user.UID.String(),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}
}

func (h *HTTPServer) createUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &NewUserParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		logging.FromContext(ctx).Errorf("unable to parse request create user body %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	nu := &domain.NewUser{
		Username: p.Username,
		Email:    p.Email,
		Password: p.Password,
	}

	u, e := h.bundle.UserService.Create(ctx, nu)
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to create new user %v", e)

		if e == domain.ErrUniqueExists {
			c := http.StatusConflict
			http.Error(w, http.StatusText(c), c)
		}

		c := http.StatusInternalServerError
		http.Error(w, http.StatusText(c), c)
		return
	}

	w.WriteHeader(http.StatusCreated)
	e = json.NewEncoder(w).Encode(newUserResponse(u))
	if e != nil {
		c := http.StatusInternalServerError
		logging.FromContext(ctx).Errorf("failed to encode user %v", e)
		http.Error(w, http.StatusText(c), c)
	}
}

// FetchUserResponse http fetch user response model
type FetchUserResponse struct {
	User *UserPayload `json:"user"`
}

func newFetchUserResponse(user *domain.User) *FetchUserResponse {
	return &FetchUserResponse{
		User: &UserPayload{
			UID:       user.UID.String(),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}
}

func (h *HTTPServer) fetchUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	sID := chi.URLParam(r, "user_id")
	uid, e := uuid.Parse(sID)
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to parse request fetch user parameters %v", e)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// verify jwt token matches user id scope
	sid := ctx.Value(mw.User)
	if sid != uid.String() {
		http.Error(w, "token claims do not match user scope", http.StatusUnauthorized)
		return
	}

	u, e := h.bundle.UserService.Read(ctx, uid)
	if e != nil {

		logging.FromContext(ctx).Errorf("failed to read user %v", e)

		if e == domain.ErrResourceNotFound {
			c := http.StatusNotFound
			http.Error(w, http.StatusText(c), c)
		}

		c := http.StatusInternalServerError
		http.Error(w, http.StatusText(c), c)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(newFetchUserResponse(u))
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
		return
	}
}

// UserLoginRequest http user login request model
type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Bind parse user login request from http
func (ulp *UserLoginRequest) Bind(_ *http.Request) error {

	if len(ulp.Username) < 1 {
		return errors.New("invalid username provided")
	} else if len(ulp.Password) < 1 {
		return errors.New("invalid password provided")
	}

	return nil
}

// UserLoginResponse http user login response model
type UserLoginResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"access_token"`
}

func (h *HTTPServer) userLogin(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &UserLoginRequest{}
	e := render.Bind(r, p)
	if e != nil {
		http.Error(w, "invalid request object", http.StatusBadRequest)
		return
	}

	uid, e := h.bundle.UserService.Login(ctx, p.Username, p.Password)
	if e != nil {
		logging.FromContext(ctx).Errorf("failed to check user login %v", e)
		http.Error(w, "unable to verify user login", http.StatusInternalServerError)
		return
	}

	c := &mw.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "test",
			Subject:   "somebody",
			ID:        "1",
			Audience:  []string{"somebody_else"},
		},
		UserID: uid,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodES384, c)

	s, e := t.SignedString(h.privateKey)
	if e != nil {
		logging.FromContext(ctx).Errorf("failed to sign jwt token %v", e)
		http.Error(w, "failed to generate jwt token", http.StatusInternalServerError)
		return
	}

	rsp := &UserLoginResponse{
		UserID: uid,
		Token:  s,
	}

	w.WriteHeader(http.StatusAccepted)
	if e := json.NewEncoder(w).Encode(rsp); e != nil {
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
		return
	}
}

// AddContactPayload http add contact request model
type AddContactPayload struct {
	Owner     string `json:"owner"`
	Recipient string `json:"recipient"`
}

// AddContactParams add contact request model
type AddContactParams struct {
	NewContact *AddContactPayload `json:"new_contact"`
}

// Bind parse request into new contact model
func (adp *AddContactParams) Bind(_ *http.Request) error {

	if len(adp.NewContact.Owner) < 36 {
		return errors.New("invalid owner parameter")
	} else if len(adp.NewContact.Recipient) < 36 {
		return errors.New("invalid recipient parameter")
	}

	return nil
}

// ContactPayload http contact model
type ContactPayload struct {
	UUID      string `json:"uid"`
	Owner     string `json:"owner"`
	Recipient string `json:"recipient"`
}

// AddContactResponse add contact response model
type AddContactResponse struct {
	Payload *ContactPayload `json:"contact"`
}

func newAddContactResponse(contact *domain.Contact) *AddContactResponse {
	return &AddContactResponse{
		Payload: &ContactPayload{
			UUID:      contact.UID.String(),
			Owner:     contact.Owner.String(),
			Recipient: contact.Recipient.String(),
		},
	}
}

func (h *HTTPServer) addContact(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &AddContactParams{}
	e := render.Bind(r, p)
	if e != nil {
		logging.FromContext(ctx).Errorf("failed to decode request body %v", e)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	oID, e := uuid.Parse(p.NewContact.Owner)
	if e != nil {
		logging.FromContext(ctx).Errorf("failed to parse owner uuid %v", e)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rID, e := uuid.Parse(p.NewContact.Recipient)
	if e != nil {
		logging.FromContext(ctx).Errorf("failed to parse recipient uuid %v", e)
		http.Error(w, "invalid request body", http.StatusBadRequest)
	}

	c, e := h.bundle.ContactService.Create(ctx, &domain.NewContact{
		Owner:     oID,
		Recipient: rID,
	})
	if e != nil {
		logging.FromContext(ctx).Errorf("failed to add new contact %v", e)
		http.Error(w, "failed to add new contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	e = json.NewEncoder(w).Encode(newAddContactResponse(c))
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateUserPayload http update user model
type UpdateUserPayload struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UpdateUserParams update user request parameters
type UpdateUserParams struct {
	*UpdateUserPayload `json:"update_user"`
}

// Bind parse request into update user model
func (uup *UpdateUserParams) Bind(_ *http.Request) error {

	if uup.UpdateUserPayload == nil {
		return errors.New("missing update user params")
	}

	if uup.Username == "" {
		return errors.New("no name parameter provided")
	} else if uup.Email == "" {
		return errors.New("no email parameter provided")
	} else if uup.UID == "" {
		return errors.New("no uid parameter provided")
	}

	return nil
}

// UpdateUserResponse http update user response model
type UpdateUserResponse struct {
	User *UserPayload `json:"user"`
}

func newUpdateUserResponse(user *domain.User) *UpdateUserResponse {
	return &UpdateUserResponse{
		User: &UserPayload{
			UID:       user.UID.String(),
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}
}

func (h *HTTPServer) updateUser(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &UpdateUserParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		logging.FromContext(ctx).Errorf("failed to bind request update user to body %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	uid, e := uuid.Parse(p.UID)
	if e != nil {
		c := http.StatusBadRequest
		http.Error(w, http.StatusText(c), c)
		return
	}

	// verify jwt token claims match user scope
	sid := ctx.Value(mw.User)
	if sid != uid.String() {
		http.Error(w, "token claims do not match user scope", http.StatusUnauthorized)
		return
	}

	u, e := h.bundle.UserService.Update(ctx, &domain.UpdateUser{
		UID:      uid,
		Username: p.Username,
		Email:    p.Email,
	})
	if e != nil {

		logging.FromContext(ctx).Errorf("failed to update user %v", e)

		if e == domain.ErrUniqueExists {
			c := http.StatusConflict
			http.Error(w, http.StatusText(c), c)
			return
		} else if e == domain.ErrResourceNotFound {
			c := http.StatusNotFound
			http.Error(w, http.StatusText(c), c)
			return
		}

		c := http.StatusInternalServerError
		http.Error(w, http.StatusText(c), c)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(newUpdateUserResponse(u))
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// SearchUsersParams search user parameter model
type SearchUsersParams struct {
	Search string
}

// Bind parse request into search users params model
func (sup *SearchUsersParams) Bind(r *http.Request) error {

	sup.Search = chi.URLParam(r, "search_str")
	if sup.Search == "" {
		return errors.New("empty search parameter provided")
	}

	return nil
}

// SearchUsersResponse search user http response model
type SearchUsersResponse struct {
	UsersDetails []*UserDetailsPayload `json:"users"`
}

func newSearchUsersResponse(details []*domain.UserDetails) *SearchUsersResponse {

	uds := make([]*UserDetailsPayload, 0, len(details))

	for _, d := range details {
		uds = append(uds, &UserDetailsPayload{
			UID:      d.UUID.String(),
			Username: d.Username,
		})
	}

	return &SearchUsersResponse{
		UsersDetails: uds,
	}
}

func (h *HTTPServer) searchUsers(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &SearchUsersParams{}
	e := render.Bind(r, p)
	if e != nil {
		logging.FromContext(ctx).Errorf("invaid parameters for search users request %v", e)
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	uds, e := h.bundle.UserService.Search(ctx, p.Search)
	if e != nil {
		logging.FromContext(ctx).Errorf("failed to search users %v", e)

		if e == domain.ErrEmptyResult {

			c := http.StatusNoContent

			w.WriteHeader(c)
			e = json.NewEncoder(w).Encode(newSearchUsersResponse([]*domain.UserDetails{}))
			if e != nil {
				c = http.StatusInternalServerError
				logging.FromContext(ctx).Errorf("unable to encode response %v", e)
				http.Error(w, http.StatusText(c), c)
				return
			}
		}

		c := http.StatusInternalServerError
		http.Error(w, http.StatusText(c), c)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(newSearchUsersResponse(uds))
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}

// SearchContactsParams http search contacts params model
type SearchContactsParams struct {
	Search string
}

// Bind parse http request into search contacts params model
func (scp *SearchContactsParams) Bind(r *http.Request) error {

	ss := chi.URLParam(r, "search_str")
	if ss == "" {
		return errors.New("invalid search string parameter")
	}

	scp.Search = ss

	return nil
}

// SearchContactsResponse http search contacts response model
type SearchContactsResponse struct {
	ContactsPayload []*ContactPayload `json:"contacts"`
}

func newSearchContactsResponse(contacts []*domain.Contact) *SearchContactsResponse {

	hcs := make([]*ContactPayload, 0, len(contacts))

	for _, c := range contacts {
		hcs = append(hcs, &ContactPayload{
			UUID:      c.UID.String(),
			Owner:     c.Owner.String(),
			Recipient: c.Recipient.String(),
		})
	}

	return &SearchContactsResponse{
		ContactsPayload: hcs,
	}
}

func (h *HTTPServer) searchContacts(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &SearchContactsParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		http.Error(w, http.StatusText(c), c)
		return
	}

	cs, e := h.bundle.ContactService.Search(ctx, p.Search)
	if e != nil {
		c := http.StatusInternalServerError
		logging.FromContext(ctx).Errorf("failed to seach contacts %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(newSearchContactsResponse(cs))
	if e != nil {
		c := http.StatusInternalServerError
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, http.StatusText(c), c)
	}
}

// FetchContactParams http fetch contact params model
type FetchContactParams struct {
	UID uuid.UUID
}

// Bind parse http request into fetch contact params model
func (fcp *FetchContactParams) Bind(r *http.Request) error {

	cID := chi.URLParam(r, "contact_id")
	if cID == "" {
		return errors.New("invalid contact id parameter provided")
	}

	uid, e := uuid.Parse(cID)
	if e != nil {
		return fmt.Errorf("unable to parse provided contact id parameter %v", e)
	}

	fcp.UID = uid

	return nil
}

// FetchContactResponse http fetch contact response model
type FetchContactResponse struct {
	*ContactPayload `json:"contact"`
}

func (h *HTTPServer) fetchContact(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &FetchContactParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		logging.FromContext(ctx).Errorf("failed to bind request to fetch contact model %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	c, e := h.bundle.ContactService.Read(r.Context(), p.UID)
	if e != nil {
		c := http.StatusInternalServerError
		logging.FromContext(ctx).Errorf("unable to read contact %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(c)
	if e != nil {
		c := http.StatusInternalServerError
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, http.StatusText(c), c)
	}
}

func (h *HTTPServer) listContacts(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	sid := r.Context().Value("user").(string)
	uid, e := uuid.Parse(sid)
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to parse request user uuid %v", e)
		http.Error(w, "invalid request header", http.StatusBadRequest)
		return
	}

	cs, e := h.bundle.ContactService.List(r.Context(), uid)
	if e != nil {
		logging.FromContext(ctx).Errorf("failed to list contacts %v", e)
		http.Error(w, "failed to list contacts", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(cs)
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}

// DeleteContactPayload expected delete contact request payload
type DeleteContactPayload struct {
	UID uuid.UUID
}

// DeleteContactParams http delete contact params model
type DeleteContactParams struct {
	*DeleteContactPayload
}

// Bind parse http request into delete contact params model
func (dcp *DeleteContactParams) Bind(r *http.Request) error {

	cID := chi.URLParam(r, "contact_id")
	if cID == "" {
		return errors.New("invalid contact id parameter provided")
	}

	uid, e := uuid.Parse(cID)
	if e != nil {
		return fmt.Errorf("unable to parse cotact id %v", e)
	}

	dcp.UID = uid

	return nil
}

// DeleteContactResponse http delete contact response model
type DeleteContactResponse struct {
	Message string `json:"message"`
}

func (h *HTTPServer) deleteContact(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &DeleteContactParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		http.Error(w, http.StatusText(c), c)
		return
	}

	e = h.bundle.ContactService.Delete(ctx, p.UID)
	if e != nil {

		if e == domain.ErrResourceNotFound {
			c := http.StatusNotFound
			http.Error(w, http.StatusText(c), c)
			return
		}

		c := http.StatusInternalServerError
		logging.FromContext(ctx).Errorf("unable to delete contact %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	rp := &DeleteContactResponse{
		Message: "success",
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(rp)
	if e != nil {
		c := http.StatusInternalServerError
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, http.StatusText(c), c)
	}
}

// CreateConversationParams http create conversation params model
type CreateConversationParams struct {
	*domain.NewConversation `json:"new_conversation"`
}

// Bind parse http request into create conversation params model
func (ccp *CreateConversationParams) Bind(_ *http.Request) error {

	if ccp.NewConversation == nil {
		return errors.New("invalid request parameters")
	}

	if len(ccp.Recipients) < 2 {
		return errors.New("invalid number of recipients for conversation")
	}

	return nil
}

// CreateConversationResponse http create conversation response model
type CreateConversationResponse struct {
	*domain.Conversation `json:"conversation"`
}

func createConversationResponse(conversation *domain.Conversation) *CreateConversationResponse {
	return &CreateConversationResponse{
		Conversation: conversation,
	}
}

func (h *HTTPServer) createConversation(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &CreateConversationParams{}
	e := render.Bind(r, p)
	if e != nil {
		logging.FromContext(ctx).Errorf("invalid request body %v", e)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	c, e := h.bundle.MessengerService.CreateConversation(ctx, p.NewConversation)
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to create new conversation %v", e)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if e := json.NewEncoder(w).Encode(createConversationResponse(c)); e != nil {
		logging.FromContext(ctx).Errorf("failed to encode response %v", e)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// ListConversationsParams http list conversations params model
type ListConversationsParams struct {
	UIDString string `json:"uid"`
	UID       uuid.UUID
}

// Bind parse request into model
func (lcp *ListConversationsParams) Bind(_ *http.Request) error {

	uid, e := uuid.Parse(lcp.UIDString)
	if e != nil {
		return fmt.Errorf("failed to pase uid from request body %v", e)
	}

	lcp.UID = uid

	return nil
}

// ConversationPayload http conversation model
type ConversationPayload struct {
	UID        string    `json:"uid"`
	Recipients []string  `json:"recipients"`
	CreatedAt  time.Time `json:"created_at"`
}

// ListConversationsResponse http list conversations response model
type ListConversationsResponse struct {
	Conversations []*ConversationPayload `json:"conversations"`
}

func newListConversationsResponse(conversations []*domain.Conversation) *ListConversationsResponse {

	cs := make([]*ConversationPayload, 0, len(conversations))

	for _, c := range conversations {

		cs = append(cs, &ConversationPayload{
			UID:       c.UID.String(),
			CreatedAt: c.CreatedAt,
		})
	}

	return &ListConversationsResponse{
		Conversations: cs,
	}
}

func (h *HTTPServer) listConversations(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	p := &ListConversationsParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		logging.FromContext(ctx).Errorf("unable to parse request parameters %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	cs, e := h.bundle.MessengerService.ListConversations(ctx, p.UID)
	if e != nil {
		logging.FromContext(ctx).Errorf("failed to list conversations %v", e)
		http.Error(w, "failed to list conversations", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(newListConversationsResponse(cs))
	if e != nil {
		logging.FromContext(ctx).Errorf("unable to encode response %v", e)
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}

func (h *HTTPServer) health(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		logging.FromContext(ctx).Errorf("error encoding health check response %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
