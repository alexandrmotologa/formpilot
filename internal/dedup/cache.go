package dedup

import (
	"strings"
	"sync"

	"github.com/alexandrmotologa/formpilot/internal/storage"
)

type AccountCache struct {
	mu       sync.RWMutex
	byDomain map[string]*storage.AccountRecord
	list     []storage.AccountRecord
	db       *storage.DB
}

func NewAccountCache(db *storage.DB) (*AccountCache, error) {
	c := &AccountCache{
		byDomain: make(map[string]*storage.AccountRecord),
		db:       db,
	}

	if db != nil {
		if err := c.reload(); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *AccountCache) reload() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	accounts, err := c.db.ListAllAccounts()
	if err != nil {
		return err
	}

	c.list = accounts
	c.byDomain = make(map[string]*storage.AccountRecord)
	for i := range c.list {
		acc := &c.list[i]
		c.byDomain[strings.ToLower(acc.Domain)] = acc
	}

	return nil
}

func (c *AccountCache) GetByDomain(domain string) *storage.AccountRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.byDomain[strings.ToLower(domain)]
}

func (c *AccountCache) GetAll() []storage.AccountRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]storage.AccountRecord, len(c.list))
	copy(out, c.list)
	return out
}

func (c *AccountCache) Upsert(acc *storage.AccountRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update DB if connected
	if c.db != nil {
		if err := c.db.UpsertAccount(acc); err != nil {
			return err
		}
	}

	// Update in-memory
	normDomain := strings.ToLower(acc.Domain)
	c.byDomain[normDomain] = acc

	// Update list
	found := false
	for i := range c.list {
		if strings.EqualFold(c.list[i].Domain, acc.Domain) {
			c.list[i] = *acc
			found = true
			break
		}
	}
	if !found {
		c.list = append(c.list, *acc)
	}

	return nil
}
