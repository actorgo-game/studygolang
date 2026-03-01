// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type OftenTime time.Time

func NewOftenTime() OftenTime {
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", "2000-01-01 00:00:00", time.Local)
	return OftenTime(t)
}

func (self OftenTime) String() string {
	t := time.Time(self)
	if t.IsZero() {
		return "0000-00-00 00:00:00"
	}
	return t.Format("2006-01-02 15:04:05")
}

func (self OftenTime) MarshalBSONValue() (bsontype.Type, []byte, error) {
	t := time.Time(self)
	return bsontype.DateTime, bsoncore.AppendDateTime(nil, t.UnixMilli()), nil
}

func (this *OftenTime) UnmarshalBSONValue(bt bsontype.Type, data []byte) error {
	switch bt {
	case bsontype.DateTime:
		ms, _, ok := bsoncore.ReadDateTime(data)
		if !ok {
			return errors.New("invalid BSON DateTime data")
		}
		*this = OftenTime(time.UnixMilli(ms))
		return nil
	case bsontype.String:
		s, _, ok := bsoncore.ReadString(data)
		if !ok {
			return errors.New("invalid BSON String data")
		}
		t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
		if err != nil {
			return fmt.Errorf("OftenTime UnmarshalBSONValue parse string %q: %w", s, err)
		}
		*this = OftenTime(t)
		return nil
	case bsontype.EmbeddedDocument:
		// legacy: old data was stored as empty document {} by StructCodec
		return nil
	case bsontype.Null:
		return nil
	default:
		return fmt.Errorf("cannot decode BSON type %v into OftenTime", bt)
	}
}

func (self OftenTime) MarshalBinary() ([]byte, error) {
	return time.Time(self).MarshalBinary()
}

func (self OftenTime) MarshalJSON() ([]byte, error) {
	t := time.Time(self)
	if y := t.Year(); y < 0 || y >= 10000 {
		if y < 2000 {
			return []byte(`"2000-01-01 00:00:00"`), nil
		}
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	return []byte(t.Format(`"2006-01-02 15:04:05"`)), nil
}

func (self OftenTime) MarshalText() ([]byte, error) {
	return time.Time(self).MarshalText()
}

func (this *OftenTime) UnmarshalBinary(data []byte) error {
	var t time.Time
	if err := t.UnmarshalBinary(data); err != nil {
		return err
	}
	*this = OftenTime(t)
	return nil
}

func (this *OftenTime) UnmarshalJSON(data []byte) (err error) {
	str := string(data)
	if str == "null" {
		return nil
	}

	if str == `"0001-01-01 08:00:00"` {
		*this = NewOftenTime()
		return nil
	}

	var t time.Time
	t, err = time.ParseInLocation(`"2006-01-02 15:04:05"`, str, time.Local)
	if err != nil {
		return
	}
	*this = OftenTime(t)
	return
}

func (this *OftenTime) UnmarshalText(data []byte) (err error) {
	var t time.Time
	if err = t.UnmarshalText(data); err != nil {
		return
	}
	*this = OftenTime(t)
	return
}
