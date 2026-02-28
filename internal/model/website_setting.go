// Copyright 2017 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type DocMenu struct {
	Name string `json:"name" bson:"name"`
	Url  string `json:"url" bson:"url"`
}

type FriendLogo struct {
	Image  string `json:"image" bson:"image"`
	Url    string `json:"url" bson:"url"`
	Name   string `json:"name" bson:"name"`
	Width  string `json:"width" bson:"width"`
	Height string `json:"height" bson:"height"`
}

type FooterNav struct {
	Name      string `json:"name" bson:"name"`
	Url       string `json:"url" bson:"url"`
	OuterSite bool   `json:"outer_site" bson:"outer_site"`
}

const (
	TabRecommend = "recommend"
	TabAll       = "all"
)

type IndexNav struct {
	Tab        string           `json:"tab" bson:"tab"`
	Name       string           `json:"name" bson:"name"`
	DataSource string           `json:"data_source" bson:"data_source"`
	Children   []*IndexNavChild `json:"children" bson:"children"`
}

type IndexNavChild struct {
	Uri  string `json:"uri" bson:"uri"`
	Name string `json:"name" bson:"name"`
}

type websiteSetting struct {
	Id             int       `bson:"_id"`
	Name           string    `bson:"name"`
	Domain         string    `bson:"domain"`
	OnlyHttps      bool      `bson:"only_https"`
	TitleSuffix    string    `bson:"title_suffix"`
	Favicon        string    `bson:"favicon"`
	Logo           string    `bson:"logo"`
	StartYear      int       `bson:"start_year"`
	BlogUrl        string    `bson:"blog_url"`
	ReadingMenu    string    `bson:"reading_menu"`
	DocsMenu       string    `bson:"docs_menu"`
	Slogan         string    `bson:"slogan"`
	Beian          string    `bson:"beian"`
	FooterNav      string    `bson:"footer_nav"`
	FriendsLogo    string    `bson:"friends_logo"`
	ProjectDfLogo  string    `bson:"project_df_logo"`
	SeoKeywords    string    `bson:"seo_keywords"`
	SeoDescription string    `bson:"seo_description"`
	IndexNav       string    `bson:"index_nav"`
	CreatedAt      time.Time `bson:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at"`

	DocMenus    []*DocMenu    `bson:"-"`
	FriendLogos []*FriendLogo `bson:"-"`
	FooterNavs  []*FooterNav  `bson:"-"`
	IndexNavs   []*IndexNav   `bson:"-"`
}

var WebsiteSetting = &websiteSetting{}

func (self websiteSetting) CollectionName() string {
	return "website_setting"
}

func (this *websiteSetting) AfterLoad() {
	this.DocMenus = this.unmarshalDocsMenu()
	this.FriendLogos = this.unmarshalFriendsLogo()
	this.FooterNavs = this.unmarshalFooterNav()
	this.IndexNavs = this.unmarshalIndexNav()
}

func (this *websiteSetting) unmarshalDocsMenu() []*DocMenu {
	if this.DocsMenu == "" {
		return nil
	}

	var docMenus = []*DocMenu{}
	err := json.Unmarshal([]byte(this.DocsMenu), &docMenus)
	if err != nil {
		fmt.Println("unmarshal docs menu error:", err)
		return nil
	}

	return docMenus
}

func (this *websiteSetting) unmarshalFriendsLogo() []*FriendLogo {
	if this.FriendsLogo == "" {
		return nil
	}

	var friendLogos = []*FriendLogo{}
	err := json.Unmarshal([]byte(this.FriendsLogo), &friendLogos)
	if err != nil {
		fmt.Println("unmarshal friends logo error:", err)
		return nil
	}

	return friendLogos
}

func (this *websiteSetting) unmarshalFooterNav() []*FooterNav {
	var footerNavs = []*FooterNav{}
	err := json.Unmarshal([]byte(this.FooterNav), &footerNavs)
	if err != nil {
		fmt.Println("unmarshal footer nav error:", err)
		return nil
	}

	for _, footerNav := range footerNavs {
		if strings.HasPrefix(footerNav.Url, "/") {
			footerNav.OuterSite = false
		} else {
			footerNav.OuterSite = true
		}
	}

	return footerNavs
}

func (this *websiteSetting) unmarshalIndexNav() []*IndexNav {
	var indexNavs = []*IndexNav{}
	err := json.Unmarshal([]byte(this.IndexNav), &indexNavs)
	if err != nil {
		fmt.Println("unmarshal index nav error:", err)
		return nil
	}

	return indexNavs
}
