package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"net/http"
	"time"
)

// responseCache to save cache of one URI
type responseCache struct {
	// http
	Header http.Header
	// http status code
	Status int
	// body to Save
	Data []byte
}

// responseCache encode into bytes
func encodeCache(cache *responseCache) ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(cache); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// decode data byte into responseCache
func decodeToCache(byts []byte) (responseCache, error) {
	var (
		buffer bytes.Buffer
		c      responseCache
	)
	if _, err := buffer.Write(byts); err != nil {
		return c, err
	}

	dec := gob.NewDecoder(&buffer)
	if err := dec.Decode(&c); err != nil {
		return c, err
	}
	return c, nil
}

// cachedWriter ...
type cachedWriter struct {
	http.ResponseWriter
	cache  *responseCache
	store  Store // http status code
	status int   // key to save or get from cache
	key    string
	expire time.Duration
}

func (w cachedWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w cachedWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w cachedWriter) Write(data []byte) (int, error) {
	ret, err := w.ResponseWriter.Write(data)
	if err != nil {
		log.Printf("could not write response: %v", err)
		return ret, err
	}

	w.cache.Status = w.status
	w.cache.Header = w.Header()
	w.cache.Data = append(w.cache.Data, data...)

	// only save into store while statusCode is successful
	if w.status < 300 {
		value, err := encodeCache(w.cache)
		if err != nil {
			log.Printf("could not encode cache: %v", err)
		} else {
			if err = w.store.Set(w.key, value, w.expire); err != nil {
				log.Printf("could not set into cache: %v", err)
			}
		}
	}
	return ret, err
}
