// This code snippet is adapted from the watermill project "[github.com/ThreeDotsLabs/watermill](https://github.com/ThreeDotsLabs/watermill)".

// Original source: https://github.com/ThreeDotsLabs/watermill

package message

// Messages is a slice of messages.
type Messages []*Message

// IDs returns a slice of Messages' IDs.
func (m Messages) IDs() []string {
	ids := make([]string, len(m))

	for i, msg := range m {
		ids[i] = msg.UUID
	}

	return ids
}
