package expiration_test

import (
	"testing"
	"time"

	"git.sr.ht/~whereswaldon/forest-go"
	"git.sr.ht/~whereswaldon/forest-go/testutil"
	"git.sr.ht/~whereswaldon/forest-go/twig"

	"git.sr.ht/~athorp96/forest-ex/expiration"
)

func TestExpirationData(t *testing.T) {
	id, signer, community := testutil.MakeCommunityOrSkip(t)

	key, data, err := expiration.CreateTwigTTL(time.Minute * 5)
	md := twig.New()
	md.Values[key] = data
	bin, err := md.MarshalBinary()
	if err != nil {
		t.Errorf("should not have failed to marshal twig: %v", err)
	}

	builder := forest.Builder{User: id, Signer: signer}

	reply, err := builder.NewReply(community, "", bin)
	if err != nil {
		t.Errorf("should not have failed to construct reply: %v", err)
	}
	replyNoData, err := builder.NewReply(community, "", []byte{})
	if err != nil {
		t.Errorf("should not have failed to construct reply: %v", err)
	}

	cases := []struct {
		hasExpired bool
		hasExpiry  bool
		error
		forest.Node
	}{
		{
			hasExpired: false,
			hasExpiry:  true,
			error:      nil,
			Node:       reply,
		},
		{
			hasExpired: false,
			hasExpiry:  false,
			error:      nil,
			Node:       replyNoData,
		},
	}

	for _, testcase := range cases {
		expired, hasExpiration, err := expiration.ExpirationData(testcase.Node)
		if testcase.error != err {
			t.Errorf("expected err to be %v, got %v", testcase.error, err)
		}
		if testcase.hasExpired != expired {
			t.Errorf("expected hasExpired to be %v, got %v", testcase.hasExpired, expired)
		}
		if testcase.hasExpiry != hasExpiration {
			t.Errorf("expected hasExpiry to be %v, got %v", testcase.hasExpiry, hasExpiration)
		}
	}
}
