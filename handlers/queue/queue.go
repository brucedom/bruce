package queue

import "sync"

var (
	queueMutex   sync.Mutex
	messageQueue = make(map[string]int)  // Keeps track of message identifiers and their order
	messageOrder = make([][]byte, 0)     // Ordered list of messages
	seenMessages = make(map[string]bool) // Tracks whether a message has been added before
	incr         int                     // Incremental counter for message order
)

// Add: Adds a new message to the queue if it hasn't been seen before
func Add(d []byte) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	msgStr := string(d)
	if _, exists := seenMessages[msgStr]; !exists {
		incr++
		messageQueue[msgStr] = incr
		messageOrder = append(messageOrder, d)
		seenMessages[msgStr] = true
	}
}

// Remove: Removes a message from the queue if it exists
func Remove(d []byte) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	msgStr := string(d)
	if _, exists := messageQueue[msgStr]; exists {
		delete(messageQueue, msgStr)
		delete(seenMessages, msgStr)

		// Remove the message from messageOrder
		indexToRemove := -1
		for i, msg := range messageOrder {
			if string(msg) == msgStr {
				indexToRemove = i
				break
			}
		}
		if indexToRemove != -1 {
			messageOrder = append(messageOrder[:indexToRemove], messageOrder[indexToRemove+1:]...)
		}
	}
}

// GetNext: Retrieves the next message in the queue (in order), or nil if empty
func GetNext() []byte {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	if len(messageOrder) == 0 {
		return nil
	}

	// Return the first message in order and remove it from the queue
	nextMessage := messageOrder[0]
	messageOrder = messageOrder[1:]
	msgStr := string(nextMessage)

	// Clean up the queue for the removed message
	delete(messageQueue, msgStr)
	delete(seenMessages, msgStr)

	return nextMessage
}

// HasMessages: Checks if the queue has any messages left
func HasMessages() bool {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	return len(messageOrder) > 0
}
