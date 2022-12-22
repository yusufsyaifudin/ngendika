package mailclient

import (
	"container/list"
	"context"
	"crypto/sha1"
	"fmt"
	"sync"
	"time"
)

type ClientSmtpManager interface {
	Get(ctx context.Context, config *SmtpMailerConfig) (Client, error)
}

type ClientMngImpl struct {
	// MaxSize is the maximum number of clients allowed in the manager. When
	// this limit is reached, the least recently used client is evicted. Set
	// zero for no limit.
	MaxSize int

	// MaxAge is the maximum age of clients in the manager. Upon retrieval, if
	// a client has remained unused in the manager for this duration or longer,
	// it is evicted and nil is returned. Set zero to disable this
	// functionality.
	MaxAge time.Duration

	// Factory is the function which constructs clients if not found in the
	// manager.
	Factory func(config *SmtpMailerConfig) (*SmtpMailer, error)

	cache map[[sha1.Size]byte]*list.Element
	ll    *list.List
	mu    sync.Mutex
}

func NewClientSmtpManager() (ClientSmtpManager, error) {
	client := &ClientMngImpl{
		MaxSize: 64,
		MaxAge:  10 * time.Minute,
		Factory: NewSmtp,
		cache:   map[[sha1.Size]byte]*list.Element{},
		ll:      list.New(),
	}

	return client, nil
}

type managerItem struct {
	key      [sha1.Size]byte
	client   *SmtpMailer
	lastUsed time.Time
}

func (m *ClientMngImpl) Get(ctx context.Context, config *SmtpMailerConfig) (Client, error) {

	m.mu.Lock()
	defer m.mu.Unlock()

	key := cacheKey(config.EmailCredential)
	now := time.Now()
	if ele, exist := m.cache[key]; exist {
		item := ele.Value.(*managerItem)
		if m.MaxAge != 0 && item.lastUsed.Before(now.Add(-m.MaxAge)) {
			c, err := m.Factory(config)
			if err != nil {
				return nil, err
			}
			if c == nil {
				return nil, fmt.Errorf("cannot initate client with the credential")
			}
			item.client = c
		}
		item.lastUsed = now
		m.ll.MoveToFront(ele)
		return item.client, nil
	}

	c, err := m.Factory(config)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, fmt.Errorf("cannot initate client with the credential")
	}

	m.mu.Unlock()
	m.Add(c)
	m.mu.Lock()
	return c, nil
}

// Add adds a Client to the manager. You can use this to individually configure
// Clients in the manager.
func (m *ClientMngImpl) Add(client *SmtpMailer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := cacheKey(client.Config.EmailCredential)
	now := time.Now()
	if ele, hit := m.cache[key]; hit {
		item := ele.Value.(*managerItem)
		item.client = client
		item.lastUsed = now
		m.ll.MoveToFront(ele)
		return
	}
	ele := m.ll.PushFront(&managerItem{key, client, now})
	m.cache[key] = ele
	if m.MaxSize != 0 && m.ll.Len() > m.MaxSize {
		m.mu.Unlock()
		m.removeOldest()
		m.mu.Lock()
	}
}

// Len returns the current size of the ClientManager.
func (m *ClientMngImpl) Len() int {
	if m.cache == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ll.Len()
}

func (m *ClientMngImpl) removeOldest() {
	m.mu.Lock()
	ele := m.ll.Back()
	m.mu.Unlock()
	if ele != nil {
		m.removeElement(ele)
	}
}

func (m *ClientMngImpl) removeElement(e *list.Element) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ll.Remove(e)
	delete(m.cache, e.Value.(*managerItem).key)
}

// cacheKey is to ensure that one client with the same credential reuse the same connection.
// It uses the URL-like format: protocol://username:password@host:port
// If AuthIdentity exist then use it as query params
func cacheKey(cred *EmailCredential) [sha1.Size]byte {

	data := fmt.Sprintf("%s://%s:%s@%s:%d",
		cred.Protocol, cred.Username, cred.Password, cred.ServerHost, cred.ServerPort,
	)

	if cred.AuthIdentity != "" {
		data = fmt.Sprintf("%s?auth_identity=%s", data, cred.AuthIdentity)
	}

	return sha1.Sum([]byte(data))
}
