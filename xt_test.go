package xt_test

// ----------------------------------------------------------------------------------
// xt_test.go for Go's xt package
// Copyright 2019,2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices.  Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"log"

	"github.com/waldurbas/xt"

	"testing"
)

func Test_UUID(t *testing.T) {
	log.Println("test.UUID")
	b := xt.UUID()
	log.Println("b: ", b)
	s := string(b)
	log.Println("s: ", s)

	u := xt.UUID36(s)
	log.Println("u: ", u)
	ss := xt.StripUUID36(u)

	if ss != s {
		t.Errorf("test UUID fail..")
		return
	}
}

func Test_Encode(t *testing.T) {
	log.Println("test.Encode")
	a := []uint32{1, 0xe00000ef, 0xffffffff, 0x0a0d0c0d}

	for _, u := range a {
		b := xt.EncodeInt32(u)
		log.Printf("u: %.x, b: 0x%v\n", u, string(b))
		x := xt.DecodeInt32(b)
		if u != x {
			t.Errorf("test Encode/Decode fail..")
			return
		}
	}
}

func Test_Gzip(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 22, 87, 45}
	s := xt.Gzip(&data)
	log.Println("gzip.bytes", data)
	log.Println("gzip.bytes.coded", s)

	var dd []byte
	xt.Gunzip(&s, &dd)
	log.Println("gzip.bytes.decoded", dd)
}
