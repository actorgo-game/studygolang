// Copyright 2018 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// https://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import "time"

const (
	DLArchived = iota
	DLStable
	DLFeatured
	DLUnstable
)

// Download go 下载
type Download struct {
	Id          int       `bson:"_id"`
	Version     string    `bson:"version"`
	Filename    string    `bson:"filename"`
	Kind        string    `bson:"kind"`
	OS          string    `bson:"os"`
	Arch        string    `bson:"arch"`
	Size        int       `bson:"size"`
	Checksum    string    `bson:"checksum"`
	Category    int       `bson:"category"`
	IsRecommend bool      `bson:"is_recommend"`
	Seq         int       `bson:"seq"`
	CreatedAt   time.Time `bson:"created_at"`
}
