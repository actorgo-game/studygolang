// Copyright 2016 The StudyGolang Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	gio "io"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/studygolang/studygolang/db"
	"github.com/studygolang/studygolang/internal/model"

	"github.com/polaris1119/config"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/times"
	"github.com/qiniu/api.v6/conf"
	"github.com/qiniu/api.v6/io"
	"github.com/qiniu/api.v6/rs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MaxImageSize = 5 << 20 // 5M
)

// http://developer.qiniu.com/code/v6/sdk/go-sdk-6.html
type UploaderLogic struct {
	bucketName string

	uptoken   string
	tokenTime time.Time
	locker    sync.RWMutex
}

var DefaultUploader = &UploaderLogic{}

func (this *UploaderLogic) InitQiniu() {
	conf.ACCESS_KEY = config.ConfigFile.MustValue("qiniu", "access_key")
	conf.SECRET_KEY = config.ConfigFile.MustValue("qiniu", "secret_key")
	conf.UP_HOST = config.ConfigFile.MustValue("qiniu", "up_host", conf.UP_HOST)
	this.bucketName = config.ConfigFile.MustValue("qiniu", "bucket_name")
}

func (this *UploaderLogic) genUpToken() {
	if this.uptoken != "" && this.tokenTime.Add(45*time.Minute).After(time.Now()) {
		return
	}

	putPolicy := rs.PutPolicy{
		Scope: this.bucketName,
	}

	this.locker.Lock()
	this.uptoken = putPolicy.Token(nil)
	this.locker.Unlock()
	this.tokenTime = time.Now()
}

func (this *UploaderLogic) uploadLocalFile(localFile, key string) (err error) {
	this.genUpToken()

	var ret io.PutRet
	var extra = &io.PutExtra{}

	err = io.PutFile(nil, &ret, this.uptoken, key, localFile, extra)

	if err != nil {
		logger.Errorln("io.PutFile failed:", err)
		return
	}

	logger.Debugln(ret.Hash, ret.Key)

	return
}

func (this *UploaderLogic) uploadMemoryFile(r gio.Reader, key string, size int) (err error) {
	this.genUpToken()

	var ret io.PutRet
	var extra = &io.PutExtra{}

	err = io.Put2(nil, &ret, this.uptoken, key, r, int64(size), extra)

	if err != nil {
		logger.Errorln("io.Put failed:", err)

		errInfo := make(map[string]interface{})
		err = json.Unmarshal([]byte(err.Error()), &errInfo)
		if err != nil {
			logger.Errorln("io.Put Unmarshal failed:", err)
			return
		}

		code, ok := errInfo["code"]
		if ok && code == 614 {
			err = nil
		}

		return
	}

	logger.Debugln(ret.Hash, ret.Key)

	return
}

func (this *UploaderLogic) UploadImage(ctx context.Context, reader gio.Reader, imgDir string, buf []byte, ext string) (string, error) {
	objLogger := GetLogger(ctx)

	md5 := goutils.Md5Buf(buf)
	objImage, err := this.findImage(md5)
	if err != nil {
		objLogger.Errorln("find image:", md5, "error:", err)
		return "", err
	}

	if objImage.Pid > 0 {
		return objImage.Path, nil
	}

	path := imgDir + "/" + md5 + ext
	if err = this.uploadMemoryFile(reader, path, len(buf)); err != nil {
		return "", err
	}

	go this.saveImage(buf, path)

	return path, nil
}

// TransferUrl 将外站图片URL转为本站，如果失败，返回原图
func (this *UploaderLogic) TransferUrl(ctx context.Context, origUrl string, prefixs ...string) (string, error) {
	if origUrl == "" || strings.Contains(origUrl, WebsiteSetting.Domain) {
		return origUrl, errors.New("origin image is empty or is " + WebsiteSetting.Domain)
	}

	if !strings.HasPrefix(origUrl, "http") {
		origUrl = "https:" + origUrl
	}

	resp, err := http.Get(origUrl)
	if err != nil {
		return origUrl, errors.New("获取图片失败")
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return origUrl, errors.New("获取图片内容失败")
	}

	md5 := goutils.Md5Buf(buf)
	objImage, err := this.findImage(md5)
	if err != nil {
		logger.Errorln("find image:", md5, "error:", err)
		return origUrl, err
	}

	if objImage.Pid > 0 {
		return objImage.Path, nil
	}

	ext := filepath.Ext(origUrl)
	if ext == "" {
		contentType := http.DetectContentType(buf)
		exts, err := mime.ExtensionsByType(contentType)
		if err != nil {
			logger.Errorln("detect extension error:", err, "orig url:", origUrl)
		} else if len(exts) > 0 {
			ext = exts[0]
		}
	}

	if ext == "" && !strings.Contains("png,jpg,jpeg,gif,bmp", strings.ToLower(ext)) {
		logger.Errorln("can't fetch extension, url:", origUrl)
		return origUrl, errors.New("can't fetch extension")
	}

	prefix := times.Format("ymd")
	if len(prefixs) > 0 {
		prefix = prefixs[0]
	}
	path := prefix + "/" + md5 + ext
	reader := bytes.NewReader(buf)

	if len(buf) > MaxImageSize {
		return origUrl, errors.New("文件太大")
	}

	err = this.uploadMemoryFile(reader, path, len(buf))
	if err != nil {
		return origUrl, err
	}

	go this.saveImage(buf, path)

	return path, nil
}

func (this *UploaderLogic) findImage(md5 string) (*model.Image, error) {
	objImage := &model.Image{}
	err := db.GetCollection("image").FindOne(context.Background(), bson.M{"md5": md5}).Decode(objImage)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	return objImage, nil
}

func (this *UploaderLogic) saveImage(buf []byte, path string) {
	objImage := &model.Image{
		Path: path,
		Md5:  goutils.Md5Buf(buf),
		Size: len(buf),
	}

	reader := bytes.NewReader(buf)
	img, _, err := image.Decode(reader)
	if err != nil {
		logger.Errorln("image decode err:", err)
	} else {
		objImage.Width = img.Bounds().Dx()
		objImage.Height = img.Bounds().Dy()
	}

	id, idErr := db.NextID("image")
	if idErr != nil {
		logger.Errorln("image nextid err:", idErr)
		return
	}
	objImage.Pid = id

	_, err = db.GetCollection("image").InsertOne(context.Background(), objImage)
	if err != nil {
		logger.Errorln("image insert err:", err)
	}
}
