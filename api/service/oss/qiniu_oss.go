package oss

import (
	"bytes"
	"chatplus/core/types"
	"chatplus/utils"
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

type QinNiuOss struct {
	config    *types.QiNiuOssConfig
	mac       *qbox.Mac
	putPolicy storage.PutPolicy
	uploader  *storage.FormUploader
	manager   *storage.BucketManager
	proxyURL  string
}

func NewQiNiuOss(appConfig *types.AppConfig) QinNiuOss {
	config := &appConfig.OSS.QiNiu
	// build storage uploader
	zone, ok := storage.GetRegionByID(storage.RegionID(config.Zone))
	if !ok {
		zone = storage.ZoneHuanan
	}
	storeConfig := storage.Config{Zone: &zone}
	formUploader := storage.NewFormUploader(&storeConfig)
	// generate token
	mac := qbox.NewMac(config.AccessKey, config.AccessSecret)
	putPolicy := storage.PutPolicy{
		Scope: config.Bucket,
	}
	if config.SubDir == "" {
		config.SubDir = "gpt"
	}
	return QinNiuOss{
		config:    config,
		mac:       mac,
		putPolicy: putPolicy,
		uploader:  formUploader,
		manager:   storage.NewBucketManager(mac, &storeConfig),
		proxyURL:  appConfig.ProxyURL,
	}
}

func (s QinNiuOss) PutFile(ctx *gin.Context, name string) (File, error) {
	// 解析表单
	file, err := ctx.FormFile(name)
	if err != nil {
		return File{}, err
	}
	// 打开上传文件
	src, err := file.Open()
	if err != nil {
		return File{}, err
	}
	defer src.Close()

	fileExt := filepath.Ext(file.Filename)
	key := fmt.Sprintf("%s/%d%s", s.config.SubDir, time.Now().UnixMicro(), fileExt)
	// 上传文件
	ret := storage.PutRet{}
	extra := storage.PutExtra{}
	err = s.uploader.Put(ctx, &ret, s.putPolicy.UploadToken(s.mac), key, src, file.Size, &extra)
	if err != nil {
		return File{}, err
	}

	return File{
		Name:   file.Filename,
		ObjKey: key,
		URL:    fmt.Sprintf("%s/%s", s.config.Domain, ret.Key),
		Ext:    fileExt,
		Size:   file.Size,
	}, nil

}

func (s QinNiuOss) PutImg(imageURL string, useProxy bool) (string, error) {
	var imageData []byte
	var err error
	if useProxy {
		imageData, err = utils.DownloadImage(imageURL, s.proxyURL)
	} else {
		imageData, err = utils.DownloadImage(imageURL, "")
	}
	if err != nil {
		return "", fmt.Errorf("error with download image: %v", err)
	}
	parse, err := url.Parse(imageURL)
	if err != nil {
		return "", fmt.Errorf("error with parse image URL: %v", err)
	}
	fileExt := utils.GetImgExt(parse.Path)
	key := fmt.Sprintf("%s/%d%s", s.config.SubDir, time.Now().UnixMicro(), fileExt)
	ret := storage.PutRet{}
	extra := storage.PutExtra{}
	// 上传文件字节数据
	err = s.uploader.Put(context.Background(), &ret, s.putPolicy.UploadToken(s.mac), key, bytes.NewReader(imageData), int64(len(imageData)), &extra)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.config.Domain, ret.Key), nil
}

func (s QinNiuOss) PutBase64(base64Img string) (string, error) {
	imageData, err := base64.StdEncoding.DecodeString(base64Img)
	if err != nil {
		return "", fmt.Errorf("error decoding base64:%v", err)
	}
	objectKey := fmt.Sprintf("%s/%d.png", s.config.SubDir, time.Now().UnixMicro())
	ret := storage.PutRet{}
	extra := storage.PutExtra{}
	// 上传文件字节数据
	err = s.uploader.Put(context.Background(), &ret, s.putPolicy.UploadToken(s.mac), objectKey, bytes.NewReader(imageData), int64(len(imageData)), &extra)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.config.Domain, ret.Key), nil
}

func (s QinNiuOss) Delete(fileURL string) error {
	var objectKey string
	if strings.HasPrefix(fileURL, "http") {
		filename := filepath.Base(fileURL)
		objectKey = fmt.Sprintf("%s/%s", s.config.SubDir, filename)
	} else {
		objectKey = fileURL
	}

	return s.manager.Delete(s.config.Bucket, objectKey)
}

var _ Uploader = QinNiuOss{}
