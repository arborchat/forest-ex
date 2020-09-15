package expiration

import (
	"fmt"
	"git.sr.ht/~whereswaldon/forest-go/twig"
	"time"
)

func CreateTwigTTL(ttl time.Duration) (twig.Key, []byte, error) {
	key := twig.Key{Name: "expiration", Version: 1}
	expiration := time.Now().Add(ttl)
	data, err := expiration.MarshalBinary()
	if err != nil {
		return twig.Key{}, nil, fmt.Errorf("Error marshalling date: %v", err)
	}
	return key, data, nil
}
