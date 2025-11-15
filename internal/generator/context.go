package generator

import (
	"math/rand"
	"time"
)

// Context holds the state and metadata for data generation
type Context struct {
	// Rand is the random number generator (seeded for determinism)
	Rand *rand.Rand

	// Table and column metadata
	TableName  string
	ColumnName string
	RowIndex   int

	// Custom data storage
	data map[string]interface{}
}

// NewContext creates a new generation context with a random seed
func NewContext() *Context {
	return NewContextWithSeed(time.Now().UnixNano())
}

// NewContextWithSeed creates a new generation context with a specific seed
func NewContextWithSeed(seed int64) *Context {
	return &Context{
		Rand: rand.New(rand.NewSource(seed)),
		data: make(map[string]interface{}),
	}
}

// Set stores a custom value in the context
func (c *Context) Set(key string, value interface{}) {
	c.data[key] = value
}

// Get retrieves a custom value from the context
func (c *Context) Get(key string) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}

// Clone creates a copy of the context
func (c *Context) Clone() *Context {
	// Create new context with same random state
	newCtx := &Context{
		Rand:       rand.New(rand.NewSource(c.Rand.Int63())),
		TableName:  c.TableName,
		ColumnName: c.ColumnName,
		RowIndex:   c.RowIndex,
		data:       make(map[string]interface{}),
	}

	// Copy custom data
	for k, v := range c.data {
		newCtx.data[k] = v
	}

	return newCtx
}