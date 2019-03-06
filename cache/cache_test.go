// Package cache ... do connect to redis with RedisConfig ref to common or other where?
// declare interfaces to use cahce in common
package main

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jademperor/common/pkg/utils"
)

func Test_responseCache_Encode_Decode(t *testing.T) {
	cache := responseCache{
		Header: http.Header{
			"Content-Type": []string{"appliction/json", "xhtml"},
			"X-Real-Ip":    []string{"127.0.0.1"},
		},
		Status: 200,
		Data:   []byte("this is body bytes"),
	}

	byts, err := encodeCache(&cache)
	if err != nil {
		t.Errorf("could encode cache: %v", err)
	}

	// logf got bytes
	// t.Logf("got encode string: %s", string(byts))

	if got, err := decodeToCache(byts); err != nil {
		t.Errorf("could encode cache: %v", err)
	} else if !reflect.DeepEqual(got, cache) {
		t.Errorf("could not decode in same way: want %v, got %v", cache, got)
	} else {
		t.Logf("status: %d, data: %s", got.Status, string(got.Data))
	}
}

func Test_cachedWriter(t *testing.T) {
	var _ http.ResponseWriter = cachedWriter{}
}

func Test_urlEscape(t *testing.T) {
	type args struct {
		prefix string
		u      string
		extern []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 1",
			args: args{
				prefix: CachePluginKey,
				u:      "http://api.baidu.com/account/open/profile",
				extern: []string{"0xe17basu12v13"},
			},
			want: "plugin.cache:http%3A%2F%2Fapi.baidu.com%2Faccount%2Fopen%2Fprofile:0xe17basu12v13",
		},
		{
			name: "case 1",
			args: args{
				prefix: CachePluginKey,
				u:      "http://api.baidu.com/account/open/profile",
				extern: []string{},
			},
			want: "plugin.cache:http%3A%2F%2Fapi.baidu.com%2Faccount%2Fopen%2Fprofile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlEscape(tt.args.prefix, tt.args.u, tt.args.extern...); got != tt.want {
				t.Errorf("urlEscape() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TODO: fix this testcase
func Test_generateKey(t *testing.T) {
	getReq, err := http.NewRequest("GET", "http://baidu.com/api/account/profile?account=123123123123", nil)
	if err != nil {
		t.Errorf("could not new request: %v", err)
		t.FailNow()
	}

	t.Log(getReq.URL.String())

	type args struct {
		req           *http.Request
		serializeForm bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case 0",
			args: args{
				req:           getReq,
				serializeForm: false,
			},
			want: "plugin.cache:%2Fapi%2Faccount%2Fprofile%3Faccount%3D123123123123",
		},
		{
			name: "case 1",
			args: args{
				req:           getReq,
				serializeForm: true,
			},
			want: "plugin.cache:%2Fapi%2Faccount%2Fprofile%3Faccount%3D123123123123:6163636f756e743d313233313233313233313233",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpyReq := utils.CopyRequest(tt.args.req)
			form := utils.ParseRequestForm(cpyReq)
			t.Logf("%s, %v", tt.args.req.RequestURI, form)
			if got := generateKey(tt.args.req.RequestURI, form, tt.args.serializeForm); got != tt.want {
				t.Errorf("generateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
