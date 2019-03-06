// Package cache ... do connect to redis with RedisConfig ref to common or other where?
// declare interfaces to use cahce in common
package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/jademperor/api-proxier/plugin"
	"github.com/jademperor/common/pkg/utils"
)

func main() {}

var (
	_ plugin.Plugin = &Cache{}
)

const (
	// CachePluginKey = "plugin.cache"
	CachePluginKey = "plugin.cache"
	// CachePageKey   = "plugin.cache.page"
	CachePageKey = "plugin.cache.page"
)

// New PluginStore ...
func New(cfgData []byte) plugin.Plugin {
	c := &Cache{
		store:         NewInMemoryStore(),
		serializeForm: false,
		status:        plugin.Working,
		enabled:       true,
	}

	config := new(Config)
	if err := json.Unmarshal(cfgData, config); err != nil {
		log.Printf("json.Unmarshal(cfgData, c) got err: %v", err)
		return nil
	}

	c.Load(config.Rules)
	return c
}

// Cache to serve page cache ...
type Cache struct {
	store         Store // store interface with value
	serializeForm bool  // serialize form [query and post form]
	enabled       bool
	status        plugin.PlgStatus
	regexps       []*regexp.Regexp // store regular expression
	cntRegexp     int              // count of regexps
}

// generate a key with the given http.Request and serializeForm flag
// [done] TODO: post method URI need to be cached or not? serialize the form with URI can solve this?
func generateKey(URI string, form url.Values, serializeForm bool) string {
	var (
		formEncode string
	)
	if serializeForm {
		formEncode = utils.EncodeFormToString(form)
		return urlEscape(CachePluginKey, URI, formEncode)
	}

	return urlEscape(CachePluginKey, URI)
}

func urlEscape(prefix, u string, extern ...string) string {
	key := url.QueryEscape(u)
	if len(key) > 200 {
		h := sha1.New()
		io.WriteString(h, u)
		key = string(h.Sum(nil))
	}
	var buffer bytes.Buffer
	buffer.WriteString(prefix)
	buffer.WriteString(":")
	buffer.WriteString(key)
	for _, s := range extern {
		buffer.WriteString(":")
		buffer.WriteString(s)
	}
	return buffer.String()
}

// Load no cache rule settings, if the URI macthed any rule in rules
// then abort cache plugin processing
func (c *Cache) Load(rules []*Rule) {
	c.regexps = make([]*regexp.Regexp, len(rules))
	c.cntRegexp = 0
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		c.regexps[c.cntRegexp] = regexp.MustCompile(r.Regexp)
		c.cntRegexp++
	}

	fmt.Printf("compile regular count: %d\n", c.cntRegexp)
}

// match NocacheRule, true means no cache
// fasle means cache
func (c *Cache) matchCacheRule(uri string) bool {
	if c.cntRegexp == 0 {
		return false
	}

	var (
		checkChan = make(chan bool, c.cntRegexp)
		counter   int
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer close(checkChan)

	for _, reg := range c.regexps {
		// fmt.Printf("reg: %s matched\n", reg.String())
		go func(ctx context.Context, reg *regexp.Regexp, c chan<- bool) {
			// to catch send on close channel
			defer func() { recover() }()
			select {
			case <-ctx.Done():
				// println("timeout matchCacheRule")
				break
			default:
				c <- reg.MatchString(uri)
			}
		}(ctx, reg, checkChan)
	}

	for checked := range checkChan {
		if checked {
			fmt.Printf("check path [%s] can be cached: %v\n", uri, checked)
			return checked
		}
		// counter to mark all gorountine called finished
		counter++
		if counter >= c.cntRegexp {
			break
		}
	}
	return false
}

// Handle implement the interface Plugin
// [fixed] TOFIX: cannot set cache to response
func (c *Cache) Handle(ctx *plugin.Context) {
	defer plugin.Recover("Cache")

	if !c.matchCacheRule(ctx.Path) {
		log.Printf("plugin.Cache will not work with path: %s\n", ctx.Path)
		return
	}

	log.Println("plugin.Cache is working")
	key := generateKey(ctx.Request().URL.RequestURI(), ctx.Form, c.serializeForm)
	if c.store.Exists(key) {
		// if exists key then load from cache and then
		// write to http.ResponseWriter
		byts, err := c.store.Get(key)
		if err != nil {
			ctx.SetError(fmt.Errorf("plugin.cache Get cache err: %v", err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// decode into cache
		cache, err := decodeToCache(byts)
		if err != nil || cache.Status == 0 {
			ctx.SetError(fmt.Errorf("plugin.cache decode cache err: %v", err))
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// write to response
		// set cache to responseWriter
		// log.Println("hit cache", string(cache.Data), cache)
		ctx.ResponseWriter().WriteHeader(cache.Status)
		ctx.ResponseWriter().Write(cache.Data)
		for k, vals := range cache.Header {
			for _, v := range vals {
				ctx.ResponseWriter().Header().Set(k, v)
			}
		}
		ctx.AbortWithStatus(http.StatusOK)
		return
	}

	// continue process
	// println("does not hit cache")
	writer := cachedWriter{
		ResponseWriter: ctx.ResponseWriter(),
		cache:          &responseCache{},
		store:          c.store,
		status:         http.StatusOK,
		key:            key,
		expire:         DefaultExpire,
	}

	ctx.SetResponseWriter(writer)
	ctx.Next()
}

// Enabled ...
func (c *Cache) Enabled() bool {
	return c.enabled
}

// Status ...
func (c *Cache) Status() plugin.PlgStatus {
	return c.status
}

// Name ...
func (c *Cache) Name() string {
	return "plugin.cache"
}

// Enable ...
func (c *Cache) Enable(enabled bool) {
	c.enabled = enabled
	if !enabled {
		c.status = plugin.Stopped
	} else {
		c.status = plugin.Working
	}
}
