// Copyright (c) 2021 zacksleo <zacksleo@gmail.com>
// MIT Licence - http://opensource.org/licenses/MIT

/*
金山文档 Workflow
1. 使用kdocs.cn中的sessionid 登录
2. 查询最近的文档
3. 根据关键词查询文档
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	aw "github.com/deanishe/awgo"
)

var (
	wf              *aw.Workflow
	recentCacheName = "recent.json"     // Filename of cached repo list
	maxCacheAge     = 180 * time.Minute // How long to cache repo list for
)

//注销登录时，删除钥匙串中存储的信息，删除缓存信息
func logout() {
	wf.Keychain.Delete("wps_sid")
	wf.ClearCache()
	wf.WarnEmpty("注销成功", "")
	wf.SendFeedback()
}

// 登录成功后，设置钥匙串
func login(query string) {
	if query == "" {
		wf.WarnEmpty("账号没有登录", "请先登录: wps {wps_sid}")
		wf.SendFeedback()
		return
	}
	wf.Keychain.Set("wps_sid", query)
	log.Printf("set keychain =%s\n", query)
	getLatest(query)
}

// LatestFile struct
type LatestFile struct {
	AppType            string `json:"app_type"` //et
	B64fname           string `json:"b64fname"` //MjAyMS0wMS0zMCsxNF8xOV80Mi54bHN4
	CollectionTime     int64  `json:"collection_time"`
	Ctime              int64  `json:"ctime"`                //1612406773
	CurrentDeviceID    string `json:"current_device_id"`    //PC
	CurrentDeviceName  string `json:"current_device_name"`  //Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36
	CurrentDeviceType  string `json:"current_device_type"`  //Browser
	Deleted            int    `json:"deleted"`              //0
	FileCtime          int    `json:"file_ctime"`           //1612085658
	FileSrc            string `json:"file_src"`             //能源物联网
	FileSrcType        string `json:"file_src_type"`        //group
	FileID             string `json:"fileid"`               //103628990123
	GroupType          string `json:"group_type"`           //normal
	GroupID            int    `json:"groupid"`              // 1302153894
	IsTmp              int    `json:"is_tmp"`               // 1
	Mtime              int64  `json:"mtime"`                // 1613389439968
	Name               string `json:"name"`                 // "换电记录导出.xlsx"
	Operation          string `json:"operation"`            // "open"
	OriginalDeviceName string `json:"original_device_name"` // "fangyandeMacBook-Pro"
	OriginalDeviceType string `json:"original_device_type"` // "PC",
	Path               string `json:"path"`                 // "/Users/zacksleo/Downloads/换电记录导出.xlsx",
	RoamingID          string `json:"roamingid"`            // "86330293202",
	Size               int    `json:"size"`                 // 9917,
}

// Doc Item
type Doc struct {
	GroupID  int64       `json:"groupid"`  // 1129357989
	ParentID int64       `json:"parentid"` // 91168130221
	Fname    string      `json:"fname"`    // "钱江服务器端口规划20201030(1).xlsx"
	Fsize    int         `json:"fsize"`    // 12963,
	Ftype    string      `json:"ftype"`    // "file"
	Ctime    int64       `json:"ctime"`    //1612406773
	Mtime    int64       `json:"mtime"`    // 1613389439968
	Deleted  interface{} `json:"deleted"`  //false
	Path     string      `json:"path"`     // 我的云文档/钱江摩托车联网/2020-10 一期/部署资,
	NewPath  string      `json:"new_path"` // 我的云文档/钱江摩托车联网/2020-10 一期/部署资
	ID       int64       `json:"id"`       //91705017005
}

// QueryResult define
type QueryResult struct {
	Total  int   `json:"total"`
	Status int   `json:"status"`
	Files  []Doc `json:"files"`
}

func getTimeDiff(timestamp int64) string {
	now := time.Now()
	sec := now.Unix()
	diff := sec - timestamp
	if diff < 60 {
		return fmt.Sprintf("%d秒", diff)
	}
	if diff < 3600 {
		return fmt.Sprintf("%d分", diff/60)
	}
	if diff < 86400 {
		return fmt.Sprintf("%d小时", diff/3600)
	}
	return fmt.Sprintf("%d天", diff/86400)
}

//查询最近的文档
func getLatest(wpsSid string) {

	files := []LatestFile{}

	if wf.Cache.Exists(recentCacheName) {
		wf.Cache.LoadJSON(recentCacheName, &files)
	}

	if wf.Cache.Expired(recentCacheName, maxCacheAge) {
		req, err := http.NewRequest("GET", "https://www.kdocs.cn/3rd/drive/api/v3/roaming", nil)
		req.Header.Add("Cookie", "wps_sid="+wpsSid)
		q := req.URL.Query()
		q.Add("without_sid", "true")
		q.Add("offset", "0")
		q.Add("count", "20")
		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			wf.WarnEmpty("查询失败", err.Error())
			return
		}

		result, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(result, &files)

		wf.Cache.StoreJSON(recentCacheName, files)
	}

	for _, file := range files {
		wf.NewItem(fmt.Sprintf("%s", file.Name)).
			Subtitle(fmt.Sprintf("%s前 %s上阅读 %s", getTimeDiff(file.Mtime/1000), file.OriginalDeviceType, file.OriginalDeviceName)).
			Valid(true).
			Var("fileid", file.FileID)
	}
	wf.SendFeedback()
}

func queryDocs(wpsSid string, query string) {
	queryResult := QueryResult{}
	files := queryResult.Files
	cacheName := query + ".json"

	if wf.Cache.Exists(cacheName) {
		wf.Cache.LoadJSON(cacheName, &files)
	}

	if wf.Cache.Expired(cacheName, maxCacheAge) {
		req, err := http.NewRequest("GET", "https://www.kdocs.cn/3rd/drive/api/v3/search/files", nil)
		req.Header.Add("Cookie", "wps_sid="+wpsSid)
		q := req.URL.Query()
		q.Add("offset", "0")
		q.Add("count", "20")
		q.Add("sort_by", "mtime")
		q.Add("order", "DESC")
		q.Add("search_group_info", "true")
		q.Add("search_operator_name", "true")
		q.Add("include_device_info", "true")
		q.Add("searchname", query)
		q.Add("search_file_content", "false")
		q.Add("search_file_name", "true")
		req.URL.RawQuery = q.Encode()

		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			wf.WarnEmpty("查询失败", err.Error())
			return
		}

		result, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(result, &queryResult)
		wf.Cache.StoreJSON(cacheName, queryResult.Files)
	}

	if len(files) == 0 {
		wf.WarnEmpty("没有找到文件", "请使用其他关键词搜索")
	}

	for _, file := range files {
		wf.NewItem(fmt.Sprintf("%s", file.Fname)).
			Subtitle(fmt.Sprintf("%s前 %s %d", getTimeDiff(file.Mtime), file.Path, file.Fsize)).
			Valid(true).
			Var("fileid", fmt.Sprintf("%d", file.ID))
	}
	wf.SendFeedback()
}

func init() {
	wf = aw.New()
}

func run() {

	query := wf.Args()[0]

	if query == "logout" {
		logout()
		return
	}
	wpsSid, err := wf.Keychain.Get("wps_sid")

	// 如果没有找到钥匙串，则提醒用户登录
	if err != nil {
		login(query)
		return
	}

	// 默认搜索最近的文档
	if query == "" {
		getLatest(wpsSid)
		return
	}

	//使用关键词搜索
	queryDocs(wpsSid, query)
}

func main() {
	wf.Run(run)
}
