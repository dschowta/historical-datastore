package test

import (
	"net/url"
	"sync"
	"time"

	"github.com/influxdb/influxdb"
	"github.com/influxdb/influxdb/messaging"
)

func init() {
	// Ensure the broker matches the handler's interface.
	_ = messaging.Handler{Broker: messaging.NewBroker()}
}

// MessagingClient represents a test client for the messaging broker.
type MessagingClient struct {
	mu      sync.Mutex
	index   uint64           // highest index
	conns   []*MessagingConn // list of all connections
	dataURL url.URL          // clients data node URL

	messagesByTopicID map[uint64][]*messaging.Message // message by topic

	PublishFunc func(*messaging.Message) (uint64, error)
	ConnFunc    func(topicID uint64) influxdb.MessagingConn
}

// NewMessagingClient returns a new instance of MessagingClient.
func NewMessagingClient(dataURL url.URL) *MessagingClient {
	c := &MessagingClient{
		messagesByTopicID: make(map[uint64][]*messaging.Message),
		dataURL:           dataURL,
	}
	c.PublishFunc = c.DefaultPublishFunc
	c.ConnFunc = c.DefaultConnFunc
	return c
}

// NewMessagingClient returns a new instance of MessagingClient.
func NewDefaultMessagingClient() *MessagingClient {
	testDataURL, _ := url.Parse("http://localhost:1234/data")
	return NewMessagingClient(*testDataURL)
}

func (c *MessagingClient) Open(path string) error { return nil }

// Close closes all open connections.
func (c *MessagingClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, conn := range c.conns {
		conn.Close()
	}
	c.conns = nil

	return nil
}

func (c *MessagingClient) URLs() []url.URL   { return []url.URL{{Host: "local"}} }
func (c *MessagingClient) SetURLs([]url.URL) {}

func (c *MessagingClient) Publish(m *messaging.Message) (uint64, error) { return c.PublishFunc(m) }

// DefaultPublishFunc sets an autoincrementing index on the message and sends it to each topic connection.
func (c *MessagingClient) DefaultPublishFunc(m *messaging.Message) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Increment index and assign it to message.
	c.index++
	m.Index = c.index

	// Append message to the topic.
	c.messagesByTopicID[m.TopicID] = append(c.messagesByTopicID[m.TopicID], m)

	// Send to each connection for the topic.
	for _, conn := range c.conns {
		if conn.topicID == m.TopicID {
			conn.Send(m)
		}
	}

	return m.Index, nil
}

func (c *MessagingClient) Conn(topicID uint64) influxdb.MessagingConn {
	return c.ConnFunc(topicID)
}

func (c *MessagingClient) CloseConn(topicID uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conns := []*MessagingConn{}
	for _, conn := range c.conns {
		if conn.topicID == topicID {
			if err := conn.Close(); err != nil {
				return err
			}
			continue
		}
		conns = append(conns, conn)
	}
	c.conns = conns
	return nil
}

// DefaultConnFunc returns a connection for a specific topic.
func (c *MessagingClient) DefaultConnFunc(topicID uint64) influxdb.MessagingConn {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create new connection.
	conn := NewMessagingConn(topicID, c.dataURL)

	// Track connections.
	c.conns = append(c.conns, conn)

	return conn
}

// Sync blocks until a given index has been sent through the client.
func (c *MessagingClient) Sync(index uint64) {
	for {
		c.mu.Lock()
		if c.index >= index {
			c.mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return
		}
		c.mu.Unlock()

		// Otherwise wait momentarily and check again.
		time.Sleep(1 * time.Millisecond)
	}
}

// MessagingConn represents a mockable connection implementing influxdb.MessagingConn.
type MessagingConn struct {
	mu      sync.Mutex
	topicID uint64
	index   uint64
	dataURL url.URL
	c       chan *messaging.Message
}

// NewMessagingConn returns a new instance of MessagingConn.
func NewMessagingConn(topicID uint64, dataURL url.URL) *MessagingConn {
	return &MessagingConn{
		topicID: topicID,
		dataURL: dataURL,
	}
}

// Open starts the stream from a given index.
func (c *MessagingConn) Open(index uint64, streaming bool) error {
	// TODO: Fill connection stream with existing messages.
	c.c = make(chan *messaging.Message, 1024)
	return nil
}

// Close closes the streaming channel.
func (c *MessagingConn) Close() error {
	close(c.c)
	return nil
}

// C returns a channel for streaming message.
func (c *MessagingConn) C() <-chan *messaging.Message { return c.c }

func (c *MessagingConn) Send(m *messaging.Message) {
	// Ignore any old messages.
	c.mu.Lock()
	if m.Index <= c.index {
		c.mu.Unlock()
		return
	}
	c.index = m.Index
	c.mu.Unlock()

	// Send message to channel.
	c.c <- m
}
