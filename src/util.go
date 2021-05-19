package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
)

func getTimeDiff(timestamp int64) string {
	now := time.Now()
	sec := now.Unix()
	diff := sec - timestamp
	if diff < 60 {
		return fmt.Sprintf("%d秒前", diff)
	}
	if diff < 3600 {
		return fmt.Sprintf("%d分前", diff/60)
	}
	if diff < 86400 {
		return fmt.Sprintf("%d小时前", diff/3600)
	}
	if diff < 2592000 {
		return fmt.Sprintf("%d天前", diff/86400)
	}
	// loc, _ := time.LoadLocation("Asia/Shanghai")
	currentTime := time.Unix(timestamp, 0)
	return currentTime.Format("2006-01-02")
}

func getMd5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

//根据文件类型和文件名称获取对应的图标
func getFtypeIcon(name string, ftype string) *aw.Icon {
	switch ftype {
	case "folder":
		return &aw.Icon{Value: "folder.png"}
	case "linkfolder":
		return &aw.Icon{Value: "linkfolder.png"}
	case "file":
		return getIcon(name)
	case "sharefile":
		return getIcon(name)
	default:
		return &aw.Icon{Value: "folder.png"}
	}
}

// 根据文件名称获取对应的图标
func getIcon(name string) *aw.Icon {
	arr := strings.Split(name, ".")
	iconDoc := &aw.Icon{Value: "doc.png"}
	iconPdf := &aw.Icon{Value: "pdf.png"}
	iconPptx := &aw.Icon{Value: "pptx.png"}
	iconTxt := &aw.Icon{Value: "txt.png"}
	iconXlsx := &aw.Icon{Value: "xlsx.png"}
	iconUnknown := &aw.Icon{Value: "unknown.png"}
	switch arr[len(arr)-1] {
	case "doc":
		return iconDoc
	case "docx":
		return iconDoc
	case "pdf":
		return iconPdf
	case "pptx":
		return iconPptx
	case "txt":
		return iconTxt
	case "csv":
		return iconXlsx
	case "xlsx":
		return iconXlsx
	case "xls":
		return iconXlsx
	default:
		return iconUnknown
	}
}

func getCacheAge(key string) time.Duration {
	expireMinute, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return 180 * time.Minute
	}
	return time.Duration(expireMinute) * time.Minute
}
