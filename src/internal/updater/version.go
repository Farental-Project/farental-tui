package updater

import (
	"strconv"
	"strings"
)

type version struct {
	major int
	minor int
	patch int
}

// parseVersion reads "Major.Minor" or "Major.Minor.Patch". A missing patch is
// zero. Anything else fails, and callers treat failure as incompatible rather
// than as a match.
func parseVersion(s string) (version, bool) {
	parts := strings.Split(strings.TrimSpace(s), ".")

	if len(parts) < 2 || len(parts) > 3 {
		return version{}, false
	}

	numbers := make([]int, 3)

	for i, p := range parts {
		n, err := strconv.Atoi(p)

		if err != nil || n < 0 {
			return version{}, false
		}

		numbers[i] = n
	}

	return version{major: numbers[0], minor: numbers[1], patch: numbers[2]}, true
}

// Compatible reports whether the client may talk to a server advertising
// serverCompat. Major and minor must match exactly; the patch is ignored.
func Compatible(clientVersion, serverCompat string) bool {
	client, ok := parseVersion(clientVersion)

	if !ok {
		return false
	}

	server, ok := parseVersion(serverCompat)

	if !ok {
		return false
	}

	return client.major == server.major && client.minor == server.minor
}

// Newer reports whether candidate is a strictly later version than current.
func Newer(candidate, current string) bool {
	c, ok := parseVersion(candidate)

	if !ok {
		return false
	}

	base, ok := parseVersion(current)

	if !ok {
		return false
	}

	switch {
	case c.major != base.major:
		return c.major > base.major
	case c.minor != base.minor:
		return c.minor > base.minor
	default:
		return c.patch > base.patch
	}
}
