package xt

// ----------------------------------------------------------------------------------
// xCode.go for Go's xt package
// Copyright 2019,2020 by Waldemar Urbas
//-----------------------------------------------------------------------------------
// This Source Code Form is subject to the terms of the 'MIT License'
// A short and simple permissive license with conditions only requiring
// preservation of copyright and license notices.  Licensed works, modifications,
// and larger works may be distributed under different terms and without source code.
// ----------------------------------------------------------------------------------

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	xguid "github.com/google/uuid"
)

// EncodeInt32 #
func EncodeInt32(k uint32) []byte {
	b := make([]byte, 4)
	s := make([]byte, 8)
	binary.BigEndian.PutUint32(b, k)

	hex.Encode(s, b)

	return s
}

// DecodeInt32 #
func DecodeInt32(s []byte) uint32 {
	b := make([]byte, 4)
	hex.Decode(b, s)
	return binary.BigEndian.Uint32(b)
}

// UUID #
func UUID() []byte {
	b := xguid.New()
	r := make([]byte, 32)
	hex.Encode(r, b[:])
	return r
}

// UUID36 #
func UUID36(uid string) string {
	suid := strings.Replace(uid, "-", "", -1)
	le := len(suid)
	if le < 32 {
		suid = suid + strings.Repeat("0", 32-le)
	}
	return fmt.Sprintf("%v-%v-%v-%v-%v", suid[0:8], suid[8:12], suid[12:16], suid[16:20], suid[20:32])
}

// StripUUID36 #
func StripUUID36(uid string) string {
	return strings.Replace(uid, "-", "", -1)
}
