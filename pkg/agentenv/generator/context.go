package generator

import (
	"context"
	"sync"
)

const (
	ContextKeyProject   = "project"
	ContextKeyVersion   = "version"
	ContextKeyAnalysis  = "analysis"
	ContextKeyChunks    = "chunks"
	ContextKeyModes     = "modes"
	ContextKeyWorkflows = "workflows"
	ContextKeyGitBranch = "git_branch"
	ContextKeyGitCommit = "git_commit"
	ContextKeyGitRemote = "git_remote"
)

type Context struct {
	mu   sync.RWMutex
	ctx  context.Context
	data map[string]interface{}
}

func newContext(ctx context.Context) *Context {
	return &Context{
		ctx:  ctx,
		data: make(map[string]interface{}),
	}
}

func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Context) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, exists := c.data[key]
	return val, exists
}

func (c *Context) GetString(key string) string {
	val, exists := c.Get(key)
	if !exists {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func (c *Context) ToMap() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}
