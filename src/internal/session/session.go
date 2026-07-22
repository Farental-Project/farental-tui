// Package session tracks the authentication state of the client against the
// server, so the app can react globally when the token expires or is revoked.
package session

// ExpiredMessage is the translation key shown to the user when the server
// rejects the auth token.
const ExpiredMessage = "Your session has expired. Please log in again."

// expired reports that the server rejected the auth token. It is only ever
// touched from the bubbletea update loop, so a plain bool is enough.
var expired bool

// Expire marks the current session as rejected by the server.
func Expire() {
	expired = true
}

// Expired peeks at the flag without consuming it.
func Expired() bool {
	return expired
}

// TakeExpired returns the flag and clears it.
func TakeExpired() bool {
	e := expired
	expired = false

	return e
}
