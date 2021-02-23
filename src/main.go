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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
)

var (
	wf              *aw.Workflow
	recentCacheName = "recent.json"     // Filename of cached repo list
	groupCacheName  = "group.json"      // Filename of cached repo list
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

// FileResult define
type FileResult struct {
	NextFilter string `json:"next_filter"`
	NextOffset int    `json:"next_offset"`
	Result     string `json:"result"`
	Files      []File
}

// GroupResult define
type GroupResult struct {
	NextFilter string `json:"next_filter"`
	NextOffset int    `json:"next_offset"`
	Result     string `json:"result"`
	Files      []Group
}

// Group struct
type Group struct {
	File
	LinkgroupID int64 `json:"linkgroupid"`
}

// File struct
type File struct {
	GroupID  int    `json:"groupid"`
	Parentid int    `json:"parentid"`
	Fname    string `json:"fname"`
	Fsize    int    `json:"fsize"`
	Ftype    string `json:"ftype"`
	Ctime    int    `json:"ctime"`
	Mtime    int64  `json:"mtime"`
	Deleted  bool   `json:"deleted"`
	ID       int64  `json:"id"`
	Store    int    `json:"store"`
	Storeid  string `json:"storeid"`
	Fver     int    `json:"fver"`
	Fsha     string `json:"fsha"`
}

// GroupFile struct
type GroupFile struct {
	GroupID     string `json:"groupid"`
	IsLinkGroup bool   `json:"islinkgroup"`
	FileID      int64  `json:"fileid"`
}

// Path 路径
type Path struct {
	Fname       string `json:"fname"`
	FileID      int64  `json:"fileid"`
	CorpID      int64  `json:"corpid"`
	GroupID     int64  `json:"groupid"`
	LinkGroupID int64  `json:"linkgroupid"`
	Type        string `json:"type"`
}

// GroupFilePathResult 文件路径
type GroupFilePathResult struct {
	GroupType string `json:"group_type"`
	Path      []Path `json:"path"`
	Result    string `json:"result"`
}

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
	return fmt.Sprintf("%s", currentTime.Format("2006-01-02"))
}

func getMd5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getGroupFile(text string) GroupFile {
	groupFile := GroupFile{}
	wf.Cache.LoadJSON(getMd5(text)+".json", &groupFile)
	return groupFile
}

// 根据文件ID 和 GroupID 查询文件路径
func getFilePath(wpsSid string, groupID int, fileID string) string {
	groupFilePathResult := GroupFilePathResult{}

	cacheName := fileID + "-path.json"
	if wf.Cache.Exists(cacheName) {
		wf.Cache.LoadJSON(cacheName, &groupFilePathResult.Path)
	}

	if wf.Cache.Expired(cacheName, maxCacheAge) {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://www.kdocs.cn/3rd/drive/api/v5/groups/%d/files/%s/path", groupID, fileID), nil)
		req.Header.Add("Cookie", "wps_sid="+wpsSid)
		client := &http.Client{}
		resp, err := client.Do(req)

		if err != nil {
			wf.WarnEmpty("查询失败", err.Error())
			return ""
		}

		result, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(result, &groupFilePathResult)

		wf.Cache.StoreJSON(cacheName, groupFilePathResult.Path)
	}
	var paths []string
	for index, path := range groupFilePathResult.Path {
		if index == 0 && path.Type == "linkfolder" {
			paths = append(paths, "团队文档")
		}
		paths = append(paths, path.Fname)
	}
	return "/" + strings.Join(paths, "/")
}

//查询最近的文档
func getLatest(wpsSid string) {
	wpsCacheDir := os.Getenv("wps_cache_dir")

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

	wf.NewItem("浏览").
		Subtitle("查看所有目录").
		Valid(true).Icon(&aw.Icon{Value: "folder.png"}).
		Var("groupid", "0").Var("fileid", "0").
		Var("path", "/").Valid(true).Autocomplete("/")

	for _, file := range files {
		quickLookURL := file.Path
		if file.Path == "" {
			quickLookURL = wpsCacheDir + getFilePath(wpsSid, file.GroupID, file.FileID)
		}
		item := wf.NewItem(file.Name).
			Subtitle(fmt.Sprintf("%s %s上阅读 %s", getTimeDiff(file.Mtime/1000), file.OriginalDeviceType, file.OriginalDeviceName)).
			Valid(true).Icon(getIcon(file.Name)).
			Var("fileid", file.FileID).Var("name", file.Name).
			Var("url", fmt.Sprintf("https://www.kdocs.cn/p/%s", file.FileID)).
			Quicklook(quickLookURL)
		item.Opt().Subtitle("复制分享连接")
		item.Cmd().Subtitle("在 WPS 中查看")
	}
	wf.SendFeedback()
}

func queryDocs(wpsSid string, query string) {
	queryResult := QueryResult{}
	cacheName := query + ".json"
	wpsCacheDir := os.Getenv("wps_cache_dir")

	if wf.Cache.Exists(cacheName) {
		wf.Cache.LoadJSON(cacheName, &queryResult.Files)
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

	if len(queryResult.Files) == 0 {
		wf.WarnEmpty("没有找到文件", "请使用其他关键词搜索")
	}

	for _, file := range queryResult.Files {
		quickLookURL := wpsCacheDir + "/" + strings.Replace(file.Path, "我的云文档", "团队文档", 1) + "/" + file.Fname
		if _, err := os.Stat(quickLookURL); os.IsNotExist(err) {
			quickLookURL = strings.Replace(quickLookURL, "/团队文档", "", 1)
		}
		url := fmt.Sprintf("https://www.kdocs.cn/p/%d", file.ID)
		if file.Ftype == "folder" {
			url = fmt.Sprintf("https://www.kdocs.cn/team/%d/%d", file.GroupID, file.ID)
		}
		item := wf.NewItem(file.Fname).
			Subtitle(fmt.Sprintf("%s %s %d", getTimeDiff(file.Mtime), file.Path, file.Fsize)).
			Valid(true).Icon(getIcon(file.Fname)).
			Quicklook(quickLookURL).
			Var("name", file.Fname).
			Var("fileid", fmt.Sprintf("%d", file.ID)).
			Var("url", url)
		item.Opt().Subtitle("复制分享连接")
		item.Cmd().Subtitle("在 WPS 中查看")
	}
	wf.SendFeedback()
}

// 查询分组
func getGroups(wpsSid string) {
	queryResult := GroupResult{}

	if wf.Cache.Exists(groupCacheName) {
		wf.Cache.LoadJSON(groupCacheName, &queryResult.Files)
	}

	if wf.Cache.Expired(groupCacheName, maxCacheAge) {
		req, err := http.NewRequest("GET", "https://www.kdocs.cn/3rd/drive/api/v5/groups/special/files", nil)
		req.Header.Add("Cookie", "wps_sid="+wpsSid)
		q := req.URL.Query()
		q.Add("linkgroup", "true")
		q.Add("include", "pic_thumbnail")
		q.Add("offset", "0")
		q.Add("count", "20")
		q.Add("orderby", "mtime")
		q.Add("order", "DESC")
		q.Add("append", "false")
		req.URL.RawQuery = q.Encode()
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			wf.WarnEmpty("查询失败", err.Error())
			return
		}
		result, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(result, &queryResult)
		wf.Cache.StoreJSON(groupCacheName, queryResult.Files)
		for _, file := range queryResult.Files {
			groupid := fmt.Sprintf("%d", file.LinkgroupID)
			isLinkGroup := true
			if groupid == "0" {
				isLinkGroup = false
				groupid = fmt.Sprintf("%d", file.GroupID)
			}
			wf.Cache.StoreJSON(getMd5("/"+file.Fname)+".json", GroupFile{GroupID: groupid, IsLinkGroup: isLinkGroup, FileID: file.ID})
		}
	}

	wf.NewItem("最近").
		Subtitle("查看最近使用的文档").
		Valid(true).Icon(&aw.Icon{Value: "recent.png"}).
		Var("groupid", "0").Var("fileid", "0").
		Var("path", "").Valid(true).Autocomplete("")
	for _, file := range queryResult.Files {
		groupid := fmt.Sprintf("%d", file.LinkgroupID)
		url := fmt.Sprintf("https://www.kdocs.cn/team/%d?folderid=%d", file.LinkgroupID, file.ID)
		if groupid == "0" {
			groupid = fmt.Sprintf("%d", file.GroupID)
			url = fmt.Sprintf("https://www.kdocs.cn/mine/%d", file.ID)
		}
		item := wf.NewItem(fmt.Sprintf("%s", file.Fname)).
			Subtitle(getTimeDiff(file.Mtime)).
			Valid(true).Icon(getFtypeIcon(file.Fname, file.Ftype)).
			Var("fileid", fmt.Sprintf("%d", file.ID)).Var("name", file.Fname).Var("groupid", groupid).
			Var("url", url).
			Var("path", "/"+file.Fname).Valid(true).Autocomplete("/" + file.Fname)
		item.Opt().Subtitle("复制分享连接")
		item.Cmd().Subtitle("在 WPS 中查看")
	}
	wf.SendFeedback()
}

func getGroupFiles(path string, wpsSid string) {

	groupID := os.Getenv("groupid")
	parentID := os.Getenv("fileid")
	parentFileID := os.Getenv("parentFileid")
	wpsCacheDir := os.Getenv("wps_cache_dir")
	isLinkGroup := true
	if groupID == "" {
		groupFile := getGroupFile(path)
		groupID = groupFile.GroupID
		isLinkGroup = groupFile.IsLinkGroup
		parentID = fmt.Sprintf("%d", groupFile.FileID)
	}

	queryResult := FileResult{}
	cacheName := fmt.Sprintf("%s-%s.json", groupID, parentID)

	if wf.Cache.Exists(cacheName) {
		wf.Cache.LoadJSON(cacheName, &queryResult.Files)
	}

	if wf.Cache.Expired(cacheName, maxCacheAge) {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://www.kdocs.cn/3rd/drive/api/v5/groups/%s/files", groupID), nil)
		req.Header.Add("Cookie", "wps_sid="+wpsSid)
		q := req.URL.Query()
		q.Add("linkgroup", "true")
		q.Add("include", "pic_thumbnail")
		q.Add("offset", "0")
		q.Add("count", "20")
		q.Add("orderby", "mtime")
		q.Add("order", "DESC")
		q.Add("append", "false")
		q.Add("parentid", parentID)
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
		for _, file := range queryResult.Files {
			wf.Cache.StoreJSON(getMd5(path+"/"+file.Fname)+".json", GroupFile{GroupID: groupID, IsLinkGroup: isLinkGroup, FileID: file.ID})
		}
	}

	paths := strings.Split(path, "/")
	currentPath := "/" + strings.Join(paths[1:len(paths)-1], "/")
	item := wf.NewItem("..").
		Subtitle("返回上级目录").
		Valid(true).Icon(&aw.Icon{Value: "back.png"}).
		Var("groupid", groupID).Var("fileid", parentFileID).
		Var("path", currentPath).Valid(true).Autocomplete(currentPath)
	item.Opt().Subtitle("复制分享连接")
	item.Cmd().Subtitle("在 WPS 中查看")

	for _, file := range queryResult.Files {
		quickLookURL := wpsCacheDir + path + "/" + file.Fname
		url := fmt.Sprintf("https://www.kdocs.cn/p/%d", file.ID)
		if isLinkGroup {
			quickLookURL = wpsCacheDir + "/团队文档" + path + "/" + file.Fname
			if strings.Contains(file.Ftype, "folder") {
				url = fmt.Sprintf("https://www.kdocs.cn/team/%d/%d", file.GroupID, file.ID)
			}
		} else {
			if strings.Contains(file.Ftype, "folder") {
				url = fmt.Sprintf("https://www.kdocs.cn/mine/%d", file.ID)
			}
		}
		item := wf.NewItem(file.Fname).
			Subtitle(getTimeDiff(file.Mtime)).
			Valid(true).Icon(getFtypeIcon(file.Fname, file.Ftype)).
			Var("groupid", groupID).Var("fileid", fmt.Sprintf("%d", file.ID)).
			Var("parentFileid", parentID).
			Var("name", file.Fname).
			Var("url", url).
			Var("path", path+"/"+file.Fname).Valid(true).Autocomplete(path + "/" + file.Fname).
			Quicklook(quickLookURL)
		item.Opt().Subtitle("复制分享连接")
		item.Cmd().Subtitle("在 WPS 中查看")
	}
	wf.SendFeedback()
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

func init() {
	wf = aw.New()
}

func run() {

	query := ""
	if len(wf.Args()) > 0 {
		query = wf.Args()[0]
	}

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
	if len(query) < 1 {
		getLatest(wpsSid)
		return
	}

	if query == "/" {
		getGroups(wpsSid)
		return
	}

	if strings.HasPrefix(query, "/") {
		getGroupFiles(query, wpsSid)
		return
	}

	//使用关键词搜索
	queryDocs(wpsSid, query)
}

func main() {
	wf.Run(run)
}