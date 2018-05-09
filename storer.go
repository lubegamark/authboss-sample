package main

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/volatiletech/authboss"
	aboauth "github.com/volatiletech/authboss/oauth2"
)

var nextUserID int

// User struct for authboss
type User struct {
	ID int

	// Non-authboss related field
	Name string

	// Auth
	Email    string
	Password string

	// Confirm
	ConfirmToken string
	Confirmed    bool

	// Lock
	AttemptCount int
	LastAttempt  time.Time
	Locked       time.Time

	// Recover
	RecoverToken       string
	RecoverTokenExpiry time.Time

	// OAuth2
	OAuth2UID          string
	OAuth2Provider     string
	OAuth2AccessToken  string
	OAuth2RefreshToken string
	OAuth2Expiry       time.Time

	// Remember is in another table
}

// This pattern is useful in real code to ensure that
// we've got the right interfaces implemented.
var (
	assertUser   = &User{}
	assertStorer = &MemStorer{}

	_ authboss.User            = assertUser
	_ authboss.AuthableUser    = assertUser
	_ authboss.ConfirmableUser = assertUser
	_ authboss.LockableUser    = assertUser
	_ authboss.RecoverableUser = assertUser

	_ authboss.CreatingServerStorer    = assertStorer
	_ authboss.ConfirmingServerStorer  = assertStorer
	_ authboss.RecoveringServerStorer  = assertStorer
	_ authboss.RememberingServerStorer = assertStorer
)

// PutPID into user
func (u *User) PutPID(pid string) { u.Email = pid }

// PutPassword into user
func (u *User) PutPassword(password string) { u.Password = password }

// PutEmail into user
func (u *User) PutEmail(email string) { u.Email = email }

// PutConfirmed into user
func (u *User) PutConfirmed(confirmed bool) { u.Confirmed = confirmed }

// PutConfirmToken into user
func (u *User) PutConfirmToken(confirmToken string) { u.ConfirmToken = confirmToken }

// PutLocked into user
func (u *User) PutLocked(locked time.Time) { u.Locked = locked }

// PutAttemptCount into user
func (u *User) PutAttemptCount(attempts int) { u.AttemptCount = attempts }

// PutLastAttempt into user
func (u *User) PutLastAttempt(last time.Time) { u.LastAttempt = last }

// PutRecoverToken into user
func (u *User) PutRecoverToken(token string) { u.RecoverToken = token }

// PutRecoverExpiry into user
func (u *User) PutRecoverExpiry(expiry time.Time) { u.RecoverTokenExpiry = expiry }

// PutOAuth2UID into user
func (u *User) PutOAuth2UID(uid string) { u.OAuth2UID = uid }

// PutOAuth2Provider into user
func (u *User) PutOAuth2Provider(provider string) { u.OAuth2Provider = provider }

// PutOAuth2AccessToken into user
func (u *User) PutOAuth2AccessToken(token string) { u.OAuth2AccessToken = token }

// PutOAuth2RefreshToken into user
func (u *User) PutOAuth2RefreshToken(refreshToken string) { u.OAuth2RefreshToken = refreshToken }

// PutOAuth2Expiry into user
func (u *User) PutOAuth2Expiry(expiry time.Time) { u.OAuth2Expiry = expiry }

// PutArbitrary into user
func (u *User) PutArbitrary(values map[string]string) {
	if n, ok := values["name"]; ok {
		u.Name = n
	}
}

// GetPID from user
func (u User) GetPID() string { return u.Email }

// GetPassword from user
func (u User) GetPassword() string { return u.Password }

// GetEmail from user
func (u User) GetEmail() string { return u.Email }

// GetConfirmed from user
func (u User) GetConfirmed() bool { return u.Confirmed }

// GetConfirmToken from user
func (u User) GetConfirmToken() string { return u.ConfirmToken }

// GetLocked from user
func (u User) GetLocked() time.Time { return u.Locked }

// GetAttemptCount from user
func (u User) GetAttemptCount() int { return u.AttemptCount }

// GetLastAttempt from user
func (u User) GetLastAttempt() time.Time { return u.LastAttempt }

// GetRecoverToken from user
func (u User) GetRecoverToken() string { return u.RecoverToken }

// GetRecoverExpiry from user
func (u User) GetRecoverExpiry() time.Time { return u.RecoverTokenExpiry }

// GetArbitrary from user
func (u User) GetArbitrary() map[string]string {
	return map[string]string{
		"name": u.Name,
	}
}

// IsOAuth2User returns true if the user was created with oauth2
func (u User) IsOAuth2User() bool { return len(u.OAuth2UID) != 0 }

// GetOAuth2UID from user
func (u User) GetOAuth2UID() (uid string) { return u.OAuth2UID }

// GetOAuth2Provider from user
func (u User) GetOAuth2Provider() (provider string) { return u.OAuth2Provider }

// GetOAuth2AccessToken from user
func (u User) GetOAuth2AccessToken() (token string) { return u.OAuth2AccessToken }

// GetOAuth2RefreshToken from user
func (u User) GetOAuth2RefreshToken() (refreshToken string) { return u.OAuth2RefreshToken }

// GetOAuth2Expiry from user
func (u User) GetOAuth2Expiry() (expiry time.Time) { return u.OAuth2Expiry }

// MemStorer stores users in memory
type MemStorer struct {
	Users  map[string]User
	Tokens map[string][]string
}

// NewMemStorer constructor
func NewMemStorer() *MemStorer {
	return &MemStorer{
		Users: map[string]User{
			"rick@councilofricks.com": User{
				ID:        1,
				Name:      "Rick",
				Password:  "$2a$10$XtW/BrS5HeYIuOCXYe8DFuInetDMdaarMUJEOg/VA/JAIDgw3l4aG", // pass = 1234
				Email:     "rick@councilofricks.com",
				Confirmed: true,
			},
		},
		Tokens: make(map[string][]string),
	}
}

// Save the user
func (m MemStorer) Save(ctx context.Context, user authboss.User) error {
	u := user.(*User)
	m.Users[u.Email] = *u

	debugln("Saved user:", u.Name)
	return nil
}

// Load the user
func (m MemStorer) Load(ctx context.Context, key string) (user authboss.User, err error) {
	// Check to see if our key is actually an oauth2 pid
	provider, uid, err := authboss.ParseOAuth2PID(key)
	if err == nil {
		for _, u := range m.Users {
			if u.OAuth2Provider == provider && u.OAuth2UID == uid {
				debugln("Loaded OAuth2 user:", u.Email)
				return &u, nil
			}
		}

		return nil, authboss.ErrUserNotFound
	}

	u, ok := m.Users[key]
	if !ok {
		return nil, authboss.ErrUserNotFound
	}

	debugln("Loaded user:", u.Name)
	return &u, nil
}

// New user creation
func (m MemStorer) New(ctx context.Context) authboss.User {
	return &User{}
}

// Create the user
func (m MemStorer) Create(ctx context.Context, user authboss.User) error {
	u := user.(*User)

	if _, ok := m.Users[u.Email]; ok {
		return authboss.ErrUserFound
	}

	debugln("Created new user:", u.Name)
	m.Users[u.Email] = *u
	return nil
}

// LoadByConfirmToken looks a user up by confirmation token
func (m MemStorer) LoadByConfirmToken(ctx context.Context, token string) (user authboss.ConfirmableUser, err error) {
	for _, v := range m.Users {
		if v.ConfirmToken == token {
			debugln("Loaded user by confirm token:", token, v.Name)
			return &v, nil
		}
	}

	return nil, authboss.ErrUserNotFound
}

// LoadByRecoverToken looks a user up by confirmation token
func (m MemStorer) LoadByRecoverToken(ctx context.Context, token string) (user authboss.RecoverableUser, err error) {
	for _, v := range m.Users {
		if v.RecoverToken == token {
			debugln("Loaded user by recover token:", token, v.Name)
			return &v, nil
		}
	}

	return nil, authboss.ErrUserNotFound
}

// AddRememberToken to a user
func (m MemStorer) AddRememberToken(pid, token string) error {
	m.Tokens[pid] = append(m.Tokens[pid], token)
	debugf("Adding rm token to %s: %s\n", pid, token)
	spew.Dump(m.Tokens)
	return nil
}

// DelRememberTokens removes all tokens for the given pid
func (m MemStorer) DelRememberTokens(pid string) error {
	delete(m.Tokens, pid)
	debugln("Deleting rm tokens from:", pid)
	spew.Dump(m.Tokens)
	return nil
}

// UseRememberToken finds the pid-token pair and deletes it.
// If the token could not be found return ErrTokenNotFound
func (m MemStorer) UseRememberToken(pid, token string) error {
	tokens, ok := m.Tokens[pid]
	if !ok {
		debugln("Failed to find rm tokens for:", pid)
		return authboss.ErrTokenNotFound
	}

	for i, tok := range tokens {
		if tok == token {
			tokens[len(tokens)-1] = tokens[i]
			m.Tokens[pid] = tokens[:len(tokens)-1]
			debugf("Used remember for %s: %s\n", pid, token)
			return nil
		}
	}

	return authboss.ErrTokenNotFound
}

// NewFromOAuth2 creates an oauth2 user (but not in the database, just a blank one to be saved later)
func (m MemStorer) NewFromOAuth2(ctx context.Context, provider string, details map[string]string) (authboss.OAuth2User, error) {
	switch provider {
	case "google":
		email := details[aboauth.OAuth2Email]

		var user *User
		if u, ok := m.Users[email]; ok {
			user = &u
		} else {
			user = &User{}
		}

		// Google OAuth2 doesn't allow us to fetch real name without more complicated API calls
		// in order to do this properly in your own app, look at replacing the authboss oauth2.GoogleUserDetails
		// method with something more thorough.
		user.Name = "Unknown"
		user.Email = details[aboauth.OAuth2Email]
		user.OAuth2UID = details[aboauth.OAuth2UID]
		user.Confirmed = true

		return user, nil
	}

	return nil, errors.Errorf("unknown provider %s", provider)
}

// SaveOAuth2 user
func (m MemStorer) SaveOAuth2(ctx context.Context, user authboss.OAuth2User) error {
	u := user.(*User)
	m.Users[u.Email] = *u

	return nil
}

/*
func (s MemStorer) PutOAuth(uid, provider string, attr authboss.Attributes) error {
	return s.Create(uid+provider, attr)
}

func (s MemStorer) GetOAuth(uid, provider string) (result interface{}, err error) {
	user, ok := s.Users[uid+provider]
	if !ok {
		return nil, authboss.ErrUserNotFound
	}

	return &user, nil
}

func (s MemStorer) AddToken(key, token string) error {
	s.Tokens[key] = append(s.Tokens[key], token)
	fmt.Println("AddToken")
	spew.Dump(s.Tokens)
	return nil
}

func (s MemStorer) DelTokens(key string) error {
	delete(s.Tokens, key)
	fmt.Println("DelTokens")
	spew.Dump(s.Tokens)
	return nil
}

func (s MemStorer) UseToken(givenKey, token string) error {
	toks, ok := s.Tokens[givenKey]
	if !ok {
		return authboss.ErrTokenNotFound
	}

	for i, tok := range toks {
		if tok == token {
			toks[i], toks[len(toks)-1] = toks[len(toks)-1], toks[i]
			s.Tokens[givenKey] = toks[:len(toks)-1]
			return nil
		}
	}

	return authboss.ErrTokenNotFound
}

func (s MemStorer) ConfirmUser(tok string) (result interface{}, err error) {
	fmt.Println("==============", tok)

	for _, u := range s.Users {
		if u.ConfirmToken == tok {
			return &u, nil
		}
	}

	return nil, authboss.ErrUserNotFound
}

func (s MemStorer) RecoverUser(rec string) (result interface{}, err error) {
	for _, u := range s.Users {
		if u.RecoverToken == rec {
			return &u, nil
		}
	}

	return nil, authboss.ErrUserNotFound
}
*/
