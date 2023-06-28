package domain

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trevatk/go-chat/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"modernc.org/sqlite"
	codes "modernc.org/sqlite/lib"
)

// NewUser application layer new user model
type NewUser struct {
	Username string
	Email    string
	Password string
}

// User application layer user model
type User struct {
	UID       uuid.UUID
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserDetails application layer user details model
type UserDetails struct {
	UUID     uuid.UUID
	Username string
}

// UpdateUser application layer update user model
type UpdateUser struct {
	UID      uuid.UUID
	Username string
	Email    string
}

// UserService user management service
type UserService struct {
	db *sql.DB
}

func newUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// Create insert new user into database
//
// this method is attempting to insert a new record into the database
// the schema has a uuid primary key and two unqiue fields for username/email
// possible errors could be a duplicate primary key or duplicate unique value
func (us *UserService) Create(ctx context.Context, newUser *NewUser) (*User, error) {

	uid := uuid.New()

	co, e := us.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("failed to get database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	p, e := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost+bcrypt.MinCost)
	if e != nil {
		return nil, fmt.Errorf("failed to generated hashed password %v", e)
	}

	to, ca := context.WithTimeout(ctx, time.Second*3)
	defer ca()

	su, e := repository.New(co).InsertUser(to, &repository.InsertUserParams{
		Uuid:    uid.String(),
		Usernm:  newUser.Username,
		Email:   newUser.Email,
		Pssword: string(p),
	})
	if e != nil {

		se, ok := e.(*sqlite.Error)
		if ok {

			c := se.Code()

			if c == codes.SQLITE_CONSTRAINT_PRIMARYKEY {
				// if uuid already exists as primary key
				// recursive call to generate new uuid before insertion

				// close connection before recursive call opens new connection
				_ = co.Close()
				return us.Create(ctx, newUser)

			} else if c == codes.SQLITE_CONSTRAINT_UNIQUE {
				// possible unqiue records are email and username
				return nil, ErrUniqueExists
			}

		}

		return nil, fmt.Errorf("error excuting insert user query %v", e)
	}

	return transformSQLUser(su), nil
}

// Login authenticate user
func (us *UserService) Login(ctx context.Context, username, password string) (string, error) {

	co, e := us.db.Conn(ctx)
	if e != nil {
		return "", fmt.Errorf("failed to pull connection pool %v", e)
	}
	defer func() { _ = co.Close() }()

	p := &repository.ReadUserLoginDetailsParams{
		Email:  username,
		Usernm: username,
	}

	to, ca := context.WithTimeout(ctx, time.Second*3)
	defer ca()

	r, e := repository.New(co).ReadUserLoginDetails(to, p)
	if e != nil {
		return "", fmt.Errorf("read user login details ")
	}

	e = bcrypt.CompareHashAndPassword([]byte(r.Pssword), []byte(password))
	if e != nil {
		return "", fmt.Errorf("failed to compare passwords %v", e)
	}

	return r.Uuid, nil
}

// Read retrieve user from database by uuid
//
// this method attempts to read by uuid
// possible errors could be the provided uuid is not found
// since the record was created we can rule out empty result
func (us *UserService) Read(ctx context.Context, uuid uuid.UUID) (*User, error) {

	co, e := us.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("failed to get database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	su, e := repository.New(co).ReadUser(ctx, uuid.String())
	if e != nil {

		if e == sql.ErrNoRows {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("error excuting read user query %v", e)
	}

	return transformSQLUser(su), nil
}

// Update update existing user record in database
//
// update method attemps to update existing records
// possible errors could be uuid is not found
// or an updated value violates an existing unique record
func (us *UserService) Update(ctx context.Context, updateUser *UpdateUser) (*User, error) {

	co, e := us.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("failed to get database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	su, e := repository.New(co).UpdateUser(ctx, &repository.UpdateUserParams{
		Uuid:   updateUser.UID.String(),
		Usernm: updateUser.Username,
		Email:  updateUser.Email,
	})
	if e != nil {

		se, ok := e.(*sqlite.Error)
		if ok {

			// no user was found for provided uuid
			if se.Code() == codes.SQLITE_NOTFOUND {
				return nil, ErrResourceNotFound
			} else if se.Code() == codes.SQLITE_CONSTRAINT_UNIQUE {
				// unqiue key violation
				return nil, ErrUniqueExists
			}
		}

		return nil, fmt.Errorf("error excuting update user query %v", e)
	}

	return transformSQLUser(su), nil
}

// Search query users by email or username
//
// this method is attempting to return one/many records
// possible errors could be general, not found, empty set
func (us *UserService) Search(ctx context.Context, search string) ([]*UserDetails, error) {

	co, e := us.db.Conn(ctx)
	if e != nil {
		return nil, fmt.Errorf("failed to get database connection from pool %v", e)
	}
	defer func() { _ = co.Close() }()

	sud, e := repository.New(co).SearchUserDetails(ctx, &repository.SearchUserDetailsParams{
		Usernm: "%" + strings.ToUpper(search) + "%",
		Email:  "%" + strings.ToUpper(search) + "%",
	})
	if e != nil {

		if e == sql.ErrNoRows {
			return nil, ErrEmptyResult
		}

		return nil, fmt.Errorf("error executing search user details %v", e)
	}

	uds := make([]*UserDetails, 0, len(sud))

	for _, ud := range sud {
		uds = append(uds, &UserDetails{
			UUID:     uuid.MustParse(ud.Uuid),
			Username: ud.Usernm,
		})
	}

	return uds, nil
}

func transformSQLUser(user *repository.User) *User {

	var u time.Time

	if user.UpdatedAt.Valid {
		u = user.UpdatedAt.Time
	} else {
		u = time.Time{}
	}

	uid := uuid.MustParse(user.Uuid)

	return &User{
		UID:       uid,
		Username:  user.Usernm,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: u,
	}
}
