package chats3

import (
	"chatFileBackend/utils/constant"
	"chatFileBackend/utils/global"
	"encoding/json"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func init() {

	// 初始化文件信息数据库
	// 初始化标签数据库
	// 初始化数据库和加入相关项目

	// 初始化对象存储
	global.Init_count.Add(1)

	go func() {
		cfg_content, err := os.ReadFile(constant.S3_config_path)
		if err != nil {
			logrus.Errorln("无法加载s3配置", err.Error())
		}
		json_str := string(cfg_content)

		// 加载s3节点配置
		value := gjson.Get(json_str, "s3_points")
		if !value.Exists() {
			logrus.Errorln("无法加载s3配置:节点信息", err.Error())
			return
		}

		var pre_s3Points []S3Point
		err = json.Unmarshal([]byte(value.Raw), &pre_s3Points)
		if err != nil {
			logrus.Errorln("无法解析s3配置:节点信息", err.Error())
			return
		}

		for idx, cfg := range pre_s3Points {
			minioClient, err := minio.New(cfg.EndPoint, &minio.Options{
				Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
				Secure: cfg.UseSSL,
			})
			if err != nil {
				logrus.Warningf("s3配置的第%v项(from 0)无法加载%v", idx, err.Error())
			} else {
				pre_s3Points[idx].Point = minioClient
				s3points = append(s3points, pre_s3Points[idx])
			}
		}

		// 加载存储桶配置
		value = gjson.Get(json_str, "bucket_name")
		// var bucket_name string
		if !value.Exists() {
			logrus.Warnln("无法加载s3配置:桶名称，使用cffiles替代", err)
			bucket_name = "cffiles"
		} else {
			bucket_name = value.String()
		}

		for _, point := range s3points {
			// 节点有效性下沉到上传和下载文件时再做
			if !point.ExistBucket(bucket_name) {
				point.MakeBucket(bucket_name)
			}
		}

		global.Init_count.Done()
	}()
}
