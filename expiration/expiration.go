package expiration

import (
	"fmt"
	"time"

	"git.sr.ht/~whereswaldon/forest-go/twig"
)

func TTLKey() twig.Key {
	return twig.Key{Name: "expiration", Version: 1}
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
