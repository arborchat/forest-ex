package expiration

import (
	"fmt"
	"time"

	"git.sr.ht/~whereswaldon/forest-go"
	"git.sr.ht/~whereswaldon/forest-go/twig"
)

// IsExpired checks whether the node has an expiration set in its twig metadata
// and, if so, whether the node has expired.
func IsExpired(node forest.Node) (bool, error) {
	expired, _, err := ExpirationData(node)
	return expired, err
}

// ExpirationData checks for the presence of an expiration metadata extension as
// well as whether the node has expired. The first return value will be true
// if the node has a set expiration in the past. The second return value
// will be true if there is an expiration date set. Any errors in parsing
// the data will be returned as the final value.
func ExpirationData(node forest.Node) (hasExpired, canExpire bool, err error) {
	data, err := node.TwigMetadata()
	if err != nil {
		return false, false, fmt.Errorf("unable to extract twig from node %s: %w", node.ID(), err)
	}
	bytes, present := data.Get(TTLKeyName, 1)
	if !present {
		return false, false, nil
	}
	expiresAt, err := UnmarshalTTL(bytes)
	if err != nil {
		return false, true, fmt.Errorf("unable to parse expiration: %w", err)
	}
	return expiresAt.Before(time.Now()), true, nil
}

// ExpiresAt returns the time at which the node expires, if any. If no expiration
// date is set, it will return the zero value for a time.Time.
func ExpiresAt(node forest.Node) (time.Time, error) {
	data, err := node.TwigMetadata()
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to extract twig from node %s: %w", node.ID(), err)
	}
	bytes, present := data.Get(TTLKeyName, 1)
	if !present {
		return time.Time{}, nil
	}
	expiresAt, err := UnmarshalTTL(bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse expiration: %w", err)
	}
	return expiresAt, nil
}

const (
	TTLKeyName = "expiration"
)

func TTLKey() twig.Key {
	return twig.Key{Name: TTLKeyName, Version: 1}
}

func CreateTwigTTL(ttl time.Duration) (twig.Key, []byte, error) {
	key := TTLKey()
	expiration := time.Now().Add(ttl)
	data, err := expiration.MarshalText()
	if err != nil {
		return twig.Key{}, nil, fmt.Errorf("Error marshalling date: %v", err)
	}
	return key, data, nil
}

func UnmarshalTTL(data []byte) (time.Time, error) {
	t := time.Now()
	if err := t.UnmarshalText(data); err != nil {
		return t, fmt.Errorf("Error parsing TTL: %v", err)
	}
	return t, nil
}
