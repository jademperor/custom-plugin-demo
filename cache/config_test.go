package main

import (
	"testing"
)

func Test_initRules(t *testing.T) {
	c := &Cache{}

	type args struct {
		rules []*Rule
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "case 1",
			args: args{
				rules: []*Rule{
					{Regexp: "^/api$", Enabled: true},
					{Regexp: "/d+", Enabled: true},
				},
			},
		},
		{
			name: "case 2",
			args: args{
				rules: []*Rule{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.Load(tt.args.rules)
			if want := len(tt.args.rules); c.cntRegexp != want {
				t.Errorf("could not initRules, not equal length: %d, want %d",
					c.cntRegexp, want)
			}
		})
	}
}

func Test_matchNoCacheRule(t *testing.T) {
	c := &Cache{}
	c.Load([]*Rule{
		{Regexp: "^/api/url$", Enabled: true},
		{Regexp: "^/api/test$", Enabled: true},
		{Regexp: "^/api/hire$", Enabled: true},
	})

	type args struct {
		uri string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case 0",
			args: args{
				uri: "/api/url",
			},
			want: true,
		},
		{
			name: "case 1",
			args: args{
				uri: "/api/hhhh/ashdak",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.matchCacheRule(tt.args.uri); got != tt.want {
				t.Errorf("matchCacheRule() = %v, want %v", got, tt.want)
			}
		})
	}
}
