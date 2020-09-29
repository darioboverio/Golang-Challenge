package sample1

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// PriceService is a service that we can use to get prices for the items
// Calls to this service are expensive (they take time)
type PriceService interface {
	GetPriceFor(itemCode string) (float64, error)
}

// TransparentCache is a cache that wraps the actual service
// The cache will remember prices we ask for, so that we don't have to wait on every call
// Cache should only return a price if it is not older than "maxAge", so that we don't get stale prices
type TransparentCache struct {
	actualPriceService PriceService
	maxAge             time.Duration
	prices             map[string]Data // Data struct instead of float64 to store information about expiration on every key
	mx                 sync.RWMutex    // reader/writer mutual exclusion lock to use over prices map
}

// Data data representation to be cached with it's proper expiration time
type Data struct {
	Value      float64
	Expiration int64
}

// NewTransparentCache transparent cache constructor
func NewTransparentCache(actualPriceService PriceService, maxAge time.Duration) *TransparentCache {
	return &TransparentCache{
		actualPriceService: actualPriceService,
		maxAge:             maxAge,
		prices:             map[string]Data{},
		mx:                 sync.RWMutex{},
	}
}

// GetPriceFor gets the price for the item, either from the cache or the actual service if it was not cached or too old
func (c *TransparentCache) GetPriceFor(itemCode string) (float64, error) {
	if price, found := c.Read(itemCode); found {
		return price, nil
	}
	price, err := c.actualPriceService.GetPriceFor(itemCode)
	if err != nil {
		return 0, fmt.Errorf("getting price from service : %v", err.Error())
	}
	c.Write(itemCode, price)
	return price, nil
}

// GetPricesFor gets the prices for several items at once, some might be found in the cache, others might not
// If any of the operations returns an error, it should return an error as well
func (c *TransparentCache) GetPricesFor(itemCodes ...string) ([]float64, error) {

	results := make([]float64, len(itemCodes))
	g, _ := errgroup.WithContext(context.Background())

	for i, itemCode := range itemCodes {

		gIdx, gItemCode := i, itemCode
		g.Go(func() error {
			price, err := c.GetPriceFor(gItemCode)
			if err == nil {
				results[gIdx] = price
			}
			return err
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

// Read concurrent safe operations on map
func (c *TransparentCache) Read(k string) (float64, bool) {
	c.mx.RLock()
	data, found := c.prices[k]
	if !found {
		c.mx.RUnlock()
		return 0, false
	}
	// if key is expired, delete it from cache and return false.
	// this is a little approach just to clean, at least, those expired keys
	// that are requested again, but the right solution would be to clean all
	// expired keys that reach a predefined TTL.
	if time.Now().UnixNano() > data.Expiration {
		delete(c.prices, k)
		c.mx.RUnlock()
		return 0, false
	}
	c.mx.RUnlock()
	return data.Value, true
}

// Write concurrent safe operations on map
func (c *TransparentCache) Write(k string, v float64) {
	c.mx.Lock()
	c.prices[k] = Data{
		Value:      v,
		Expiration: time.Now().Add(c.maxAge).UnixNano(),
	}
	c.mx.Unlock()
}
