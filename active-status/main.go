package activity

import (
	forest "git.sr.ht/~whereswaldon/forest-go"
	"git.sr.ht/~whereswaldon/forest-go/twig"
)

func Heartbeat(interval time.Duration, community *forest.Community) {
	nodeBuilder, err := c.Settings.Builder()
	if err != nil {
		log.Printf("failed acquiring node builder: %v", err)
	} else {
		author = nodeBuilder.User
		convo, err := nodeBuilder.NewReply(chosen, c.ReplyEditor.Text(), []byte{})
		if err != nil {
			log.Printf("failed creating new conversation: %v", err)
		} else {
			newReply = convo
		}
	}
}
