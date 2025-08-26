package generator

import (
	"context"
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
	c.data[key] = value
}

func (c *Context) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range c.data {
		result[k] = v
	}
	return result
}
