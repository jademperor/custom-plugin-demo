package main

import (
	"io/ioutil"
	"testing"
)

func Test_RBAC(t *testing.T) {

	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		t.Fatal(err)
	}

	plg := New(data).(*RBAC)
	// t.Log(plg.users, plg.roles, plg.permissions)

	if has, need := plg.permit("/example/id", "u1"); need && !has {
		t.Logf("%v, %v", has, need)
		t.Errorf("user[%s] have not the permission to ther url: %s", "u1", "/example/id")
		t.Fail()
	}
	if has, need := plg.permit("/example/name", "u2"); need && has {
		t.Logf("%v, %v", has, need)
		t.Errorf("user[%s] have the permission to ther url: %s", "u2", "/example/name")
		t.Fail()
	}
}
