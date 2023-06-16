// Package port exposed endpoints
package port

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"go.uber.org/zap"

	"github.com/trevatk/go-chat/internal/domain"
)

// HTTPServer exposed endpoints
type HTTPServer struct {
	log    *zap.SugaredLogger
	bundle *domain.Bundle
}

// NewHTTPServer create new http server instance
func NewHTTPServer(bundle *domain.Bundle, log *zap.Logger) *HTTPServer {
	return &HTTPServer{bundle: bundle, log: log.Named("http server").Sugar()}
}

// NewRouter create new chi implementation of http.ServeMux
func NewRouter(srv *HTTPServer) *chi.Mux {

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Route("/api/v1", func(r chi.Router) {

		r.Post("/user", srv.createUser)
		r.Get("/user/{user_id}", srv.fetchUser)
		r.Put("/user", srv.updateUser)
		r.Get("/user/search/{search_str}", srv.searchUsers)

		r.Post("/contact", srv.addContact)
		r.Get("/contact/search/{search_str}", srv.searchContacts)
		r.Get("/contact/{contact_id}", srv.fetchContact)
		r.Get("/contact", srv.listContacts)
		r.Delete("/contact/{contact_id}", srv.deleteContact)

		r.Post("/conversation", srv.createConversation)
		r.Get("/conversation/", srv.listConversations)
	})

	r.Get("/health", srv.health)

	return r
}

// NewUserPayload http new user payload model
type NewUserPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
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
	}

	return nil
}

// NewUserResponse new user response model
type NewUserResponse struct {
	*UserPayload `json:"new_user"`
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

	p := &NewUserParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		h.log.Errorf("unable to parse request create user body %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	nu := &domain.NewUser{
		Username: p.Username,
		Email:    p.Email,
	}

	u, e := h.bundle.UserService.Create(r.Context(), nu)
	if e != nil {
		h.log.Errorf("unable to create new user %v", e)

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
		h.log.Errorf("failed to encode user %v", e)
		http.Error(w, http.StatusText(c), c)
	}
}

func (h *HTTPServer) fetchUser(w http.ResponseWriter, r *http.Request) {

	sID := chi.URLParam(r, "user_id")
	UID, e := uuid.Parse(sID)
	if e != nil {
		h.log.Errorf("unable to parse request fetch user parameters %v", e)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	u, e := h.bundle.UserService.Read(r.Context(), UID)
	if e != nil {

		h.log.Errorf("failed to read user %v", e)

		if e == domain.ErrResourceNotFound {
			c := http.StatusNotFound
			http.Error(w, http.StatusText(c), c)
		}

		c := http.StatusInternalServerError
		http.Error(w, http.StatusText(c), c)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(newUserResponse(u))
	if e != nil {
		h.log.Errorf("unable to encode response %v", e)
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
	*domain.NewContact
	*AddContactPayload `json:"add_contact"`
}

// Bind parse request into new contact model
func (adp *AddContactParams) Bind(_ *http.Request) error {

	oID, e := uuid.Parse(adp.AddContactPayload.Owner)
	if e != nil {
		return errors.New("invalid owner parameter")
	}

	rID, e := uuid.Parse(adp.AddContactPayload.Recipient)
	if e != nil {
		return errors.New("invalid recipient parameter")
	}

	adp.NewContact.Owner = oID
	adp.NewContact.Recipient = rID

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
	*ContactPayload
}

func (h *HTTPServer) addContact(w http.ResponseWriter, r *http.Request) {

	p := &AddContactParams{}
	e := render.Bind(r, p)
	if e != nil {
		h.log.Errorf("failed to decode request body %v", e)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	c, e := h.bundle.ContactService.Create(r.Context(), p.NewContact)
	if e != nil {
		h.log.Errorf("failed to add new contact %v", e)
		http.Error(w, "failed to add new contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	e = json.NewEncoder(w).Encode(newContactResponse(c))
	if e != nil {
		h.log.Errorf("unable to encode response %v", e)
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

func (h *HTTPServer) updateUser(w http.ResponseWriter, r *http.Request) {

	p := &UpdateUserParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		h.log.Errorf("failed to bind request update user to body %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	uid, e := uuid.Parse(p.UID)
	if e != nil {
		c := http.StatusBadRequest
		http.Error(w, http.StatusText(c), c)
		return
	}

	uu := &domain.UpdateUser{
		UID:      uid,
		Username: p.Username,
		Email:    p.Email,
	}

	u, e := h.bundle.UserService.Update(r.Context(), uu)
	if e != nil {

		h.log.Errorf("failed to update user %v", e)

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
	e = json.NewEncoder(w).Encode(u)
	if e != nil {
		h.log.Errorf("unable to encode response %v", e)
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

	p := &SearchUsersParams{}
	e := render.Bind(r, p)
	if e != nil {
		h.log.Errorf("invaid parameters for search users request %v", e)
		http.Error(w, "invalid parameters", http.StatusBadRequest)
		return
	}

	uds, e := h.bundle.UserService.Search(r.Context(), p.Search)
	if e != nil {
		h.log.Errorf("failed to search users %v", e)

		if e == domain.ErrEmptyResult {

			c := http.StatusNoContent

			w.WriteHeader(c)
			e = json.NewEncoder(w).Encode(newSearchUsersResponse([]*domain.UserDetails{}))
			if e != nil {
				c = http.StatusInternalServerError
				h.log.Errorf("unable to encode response %v", e)
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
		h.log.Errorf("unable to encode response %v", e)
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

	p := &SearchContactsParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		http.Error(w, http.StatusText(c), c)
		return
	}

	cs, e := h.bundle.ContactService.Search(r.Context(), p.Search)
	if e != nil {
		c := http.StatusInternalServerError
		h.log.Errorf("failed to seach contacts %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(newSearchContactsResponse(cs))
	if e != nil {
		c := http.StatusInternalServerError
		h.log.Errorf("unable to encode response %v", e)
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

	p := &FetchContactParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		h.log.Errorf("failed to bind request to fetch contact model %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	c, e := h.bundle.ContactService.Read(r.Context(), p.UID)
	if e != nil {
		c := http.StatusInternalServerError
		h.log.Errorf("unable to read contact %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(c)
	if e != nil {
		c := http.StatusInternalServerError
		h.log.Errorf("unable to encode response %v", e)
		http.Error(w, http.StatusText(c), c)
	}
}

func (h *HTTPServer) listContacts(w http.ResponseWriter, r *http.Request) {

	sid := r.Context().Value("user").(string)
	uid, e := uuid.Parse(sid)
	if e != nil {
		h.log.Errorf("unable to parse request user uuid %v", e)
		http.Error(w, "invalid request header", http.StatusBadRequest)
		return
	}

	cs, e := h.bundle.ContactService.List(r.Context(), uid)
	if e != nil {
		h.log.Errorf("failed to list contacts %v", e)
		http.Error(w, "failed to list contacts", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(cs)
	if e != nil {
		h.log.Errorf("unable to encode response %v", e)
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

	p := &DeleteContactParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		http.Error(w, http.StatusText(c), c)
		return
	}

	e = h.bundle.ContactService.Delete(r.Context(), p.UID)
	if e != nil {

		if e == domain.ErrResourceNotFound {
			c := http.StatusNotFound
			http.Error(w, http.StatusText(c), c)
			return
		}

		c := http.StatusInternalServerError
		h.log.Errorf("unable to delete contact %v", e)
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
		h.log.Errorf("unable to encode response %v", e)
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

	p := &CreateConversationParams{}
	e := render.Bind(r, p)
	if e != nil {
		h.log.Errorf("invalid request body %v", e)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	c, e := h.bundle.MessengerService.CreateConversation(r.Context(), p.NewConversation)
	if e != nil {
		h.log.Errorf("unable to create new conversation %v", e)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if e := json.NewEncoder(w).Encode(createConversationResponse(c)); e != nil {
		h.log.Errorf("failed to encode response %v", e)
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

		r := make([]string, 0, len(c.Recipients))

		for _, re := range c.Recipients {
			r = append(r, re.String())
		}

		cs = append(cs, &ConversationPayload{
			UID:        c.UID.String(),
			Recipients: r,
			CreatedAt:  c.CreatedAt,
		})
	}

	return &ListConversationsResponse{
		Conversations: cs,
	}
}

func (h *HTTPServer) listConversations(w http.ResponseWriter, r *http.Request) {

	p := &ListConversationsParams{}
	e := render.Bind(r, p)
	if e != nil {
		c := http.StatusBadRequest
		h.log.Errorf("unable to parse request parameters %v", e)
		http.Error(w, http.StatusText(c), c)
		return
	}

	cs, e := h.bundle.MessengerService.ListConversations(r.Context(), p.UID)
	if e != nil {
		h.log.Errorf("failed to list conversations %v", e)
		http.Error(w, "failed to list conversations", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	e = json.NewEncoder(w).Encode(newListConversationsResponse(cs))
	if e != nil {
		h.log.Errorf("unable to encode response %v", e)
		http.Error(w, "unable to encode response", http.StatusInternalServerError)
	}
}

func (h *HTTPServer) health(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		h.log.Errorf("error encoding health check response %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func newContactResponse(contact *domain.Contact) *AddContactResponse {
	return &AddContactResponse{
		ContactPayload: &ContactPayload{
			UUID:      contact.UID.String(),
			Owner:     contact.Owner.String(),
			Recipient: contact.Recipient.String(),
		},
	}
}
