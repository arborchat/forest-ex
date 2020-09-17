package activeStatus

import (
	"fmt"
	"log"
	"strconv"
	"time"

	expiration "git.sr.ht/~athorp96/forest-ex/expiration"
	forest "git.sr.ht/~whereswaldon/forest-go"
	fields "git.sr.ht/~whereswaldon/forest-go/fields"
	"git.sr.ht/~whereswaldon/forest-go/store"
	"git.sr.ht/~whereswaldon/forest-go/twig"
)

type ActiveStatus int

const (
	Active ActiveStatus = iota
	Inactive
)

// MarshalBinary translates an ActiveStatus into a byte slice
func (s *ActiveStatus) MarshalBinary() []byte {
	return []byte(strconv.Itoa(int(*s)))
}

// UnmarshalBinary creates an ActiveStatus from a byte slice
func UnmarshalBinary(b []byte) (ActiveStatus, error) {
	i, e := strconv.Atoi(string(b))
	return ActiveStatus(i), e
}

type StatusState struct {
	Status     ActiveStatus
	Creation   time.Time
	Expiration time.Time
}

// StatusManager maps users to their current status. We may eventually
// want to store more metadata such as `last active`, but for now
// just knowing is a given user is active is enough
type StatusManager struct {
	activeUsers map[string]StatusState
}

// NewStatusManager instantiates an empty StatusManager struct and returns
// a pointer to the new object.
func NewStatusManager() *StatusManager {
	mgr := StatusManager{
		activeUsers: make(map[string]StatusState),
	}
	go func() {
		cleanUpDuration := time.Second * 30
		cleanUpTicker := time.NewTicker(cleanUpDuration)
		for range cleanUpTicker.C {
			for u, s := range mgr.activeUsers {
				// If the expiration has passed,
				if time.Now().After(s.Expiration) {
					s.Status = Inactive
					mgr.activeUsers[u] = s
				}
			}
		}
	}()

	return &mgr
}

// HandleNode takes as an argument a reply node. If it is an active status message,
// it updates the StatusManager accordingly.
func (self *StatusManager) HandleNode(node forest.Node) {

	md, err := node.TwigMetadata()
	if err != nil {
		log.Printf("Error unmarshalling node %s's twig metadata: %v", node.ID(), err)
		return
	}

	twigKey := ActiveStatusKey()
	data, isActivityNode := md.Values[twigKey]
	if !isActivityNode {
		// Not activityNode
		return
	}
	status, err := UnmarshalBinary(data)
	if err != nil {
		log.Print("Malformed status request. Twig data: %b. Error: %v", data, err)
		return
	}
	log.Printf("User %v updated status to %v", node.AuthorID(), status)

	ttl, ttlExists := md.Values[expiration.TTLKey()]
	if !ttlExists {
		log.Printf("Malformed status request. Node %s has no TTL", node.ID())
		return
	}

	expireTime, err := expiration.UnmarshalTTL(ttl)
	if err != nil {
		log.Print("Malformed status request. Twig data: %b. Error: %v", data, err)
		return
	}
	self.setStatus(*node.AuthorID(), status, expireTime, node.CreatedAt())
}

// setStatus is intentionally left private so the status will always be set according to
// the logic in HandleNode
func (self *StatusManager) setStatus(user fields.QualifiedHash, status ActiveStatus, expiresAt time.Time, creation time.Time) {
	userHash := string(user.Blob)
	if time.Now().After(expiresAt) || self.activeUsers[userHash].Creation.After(creation) {
		return
	}

	state := StatusState{
		Status:     status,
		Creation:   creation,
		Expiration: expiresAt,
	}
	self.activeUsers[string(user.Blob)] = state
}

// Status returns the active status of a given user. If that user
// has never been registered by the StatusManager, they are considered
// inactive.
func (self *StatusManager) Status(user fields.QualifiedHash) ActiveStatus {
	status := Inactive

	if knownStatus, present := self.activeUsers[string(user.Blob)]; present {
		status = knownStatus.Status
	}

	return status
}

// IsActive returns whether or not a given user is listed as currently
// active. If the user has never been registered by the StatusManager,
// they are considered inactive.
func (self *StatusManager) IsActive(user fields.QualifiedHash) bool {
	return self.Status(user) == Active
}

// ActiveStatusKey defines the key used in the activeStatus metadata.
// Anywhere that references an activeStatus key must call this function
func ActiveStatusKey() twig.Key {
	return twig.Key{Name: "activity", Version: 1}
}

// activityMetadata determines the format of the twig metadata used to
// establish a node as an activity node
func activeStatusMetadata(status ActiveStatus) (twig.Key, []byte) {
	return ActiveStatusKey(), status.MarshalBinary()
}

// ActivityMetadata creates an acitivity status twig data object for
// a given status.
//
// example:
// ```
//	// Set this node to be a "activity-status" node that lives for five hours
//	ttl = time.Hour * 5
//	activityMetadata = NewActivityMetadata(Active, ttl)
//	data, _ := activityMetadata.MarshalBinary()
//	statusNode = forest.NewReply(parent, "", data)
// ```
func NewActivityMetadata(status ActiveStatus, ttl time.Duration) (*twig.Data, error) {
	data := twig.New()

	statusKey, statusData := activeStatusMetadata(status)
	ttlKey, ttlData, err := expiration.CreateTwigTTL(ttl)
	if err != nil {
		return nil, fmt.Errorf("Error creating TTL twig data: %v", err)
	}

	_, _ = data.Set("invisible", 1, []byte("true"))

	data.Values[statusKey] = statusData
	data.Values[ttlKey] = ttlData

	return data, nil
}

func NewActivityNode(statusConversation *forest.Community, builder *forest.Builder, status ActiveStatus, ttl time.Duration) (forest.Node, error) {
	md, err := NewActivityMetadata(status, ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity metadata: %v", err)
	}

	statusBlob, err := md.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal activity metadata: %v", err)
	}

	statusNode, err := builder.NewReply(statusConversation, "", statusBlob)
	if err != nil {
		return nil, fmt.Errorf("failed creating status node: %v", err)
	}

	return statusNode, nil
}

// StartActivityHeartBeat sends an activity message to a number of communities
// every time a given duration passes. It acts as a heartbeat, letting the
// communities know a user is currently connected.
func StartActivityHeartBeat(msgStore store.ExtendedStore, communities []*forest.Community, builder *forest.Builder, interval time.Duration) {
	log.Printf("Starting activity heartbeat")
	ticker := time.NewTicker(interval)
	emitHeartBeat := func() {
		for _, c := range communities {
			statusNode, err := NewActivityNode(c, builder, Active, interval)
			if err != nil {
				log.Printf("Error creating active-status node: %v", err)
			}
			err = msgStore.Add(statusNode)
			if err != nil {
				log.Printf("Error adding active status node to store: %v", err)
			}
			log.Printf("Emitted status node %s with TTL %s", statusNode.ID(), interval)
		}
	}

	// Emit an initial heartbeat
	emitHeartBeat()
	for range ticker.C {
		emitHeartBeat()
	}
}

func KillActivityHeartBeat(msgStore store.ExtendedStore, communities []*forest.Community, builder *forest.Builder) {
	ttl := time.Hour * 1
	for _, c := range communities {
		statusNode, err := NewActivityNode(c, builder, Inactive, ttl)
		if err != nil {
			log.Printf("Error creating inactive-status node: %v", err)
		}
		err = msgStore.Add(statusNode)
		if err != nil {
			log.Printf("Error adding inactive status node to store: %v", err)
		}
		log.Printf("Emitted inactive status node %s with TTL %s", statusNode.ID(), ttl)
	}
}
