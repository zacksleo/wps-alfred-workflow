package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Wps struct {
	sid string //wps_sid
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

// Path 路径
type Path struct {
	Fname       string `json:"fname"`
	FileID      int64  `json:"fileid"`
	CorpID      int64  `json:"corpid"`
	GroupID     int64  `json:"groupid"`
	LinkGroupID int64  `json:"linkgroupid"`
	Type        string `json:"type"`
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

func NewWps(sid string) *Wps {
	return &Wps{sid}
}

// GroupFilePathResult 文件路径
type GroupFilePathResult struct {
	GroupType string `json:"group_type"`
	Path      []Path `json:"path"`
	Result    string `json:"result"`
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

//查询最近的文档
func (wps *Wps) GetLatest() (*[]LatestFile, error) {
	files := []LatestFile{}

	req, _ := http.NewRequest("GET", "https://www.kdocs.cn/3rd/drive/api/v3/roaming", nil)
	req.Header.Add("Cookie", "wps_sid="+wps.sid)
	q := req.URL.Query()
	q.Add("without_sid", "true")
	q.Add("offset", "0")
	q.Add("count", "20")
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return &files, err
	}

	result, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(result, &files)
	return &files, err
}

//使用关键词查询文件
func (wps *Wps) QueryDocs(query string) (*QueryResult, error) {
	queryResult := QueryResult{}

	req, _ := http.NewRequest("GET", "https://www.kdocs.cn/3rd/drive/api/v3/search/files", nil)
	req.Header.Add("Cookie", "wps_sid="+wps.sid)
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
		return &queryResult, err
	}

	result, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(result, &queryResult)

	return &queryResult, err
}

func (wps *Wps) GetGroupFiles(groupID string, parentID string) (*FileResult, error) {

	queryResult := FileResult{}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.kdocs.cn/3rd/drive/api/v5/groups/%s/files", groupID), nil)
	req.Header.Add("Cookie", "wps_sid="+wps.sid)
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
		return &queryResult, err
	}

	result, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(result, &queryResult)

	return &queryResult, err
}

// 根据文件ID 和 GroupID 查询文件路径
func (wps *Wps) GetFilePath(groupID int, fileID string) (*GroupFilePathResult, error) {
	groupFilePathResult := GroupFilePathResult{}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.kdocs.cn/3rd/drive/api/v5/groups/%d/files/%s/path", groupID, fileID), nil)
	req.Header.Add("Cookie", "wps_sid="+wps.sid)
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return &groupFilePathResult, err
	}

	result, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(result, &groupFilePathResult)

	return &groupFilePathResult, err
}

// 查询分组
func (wps *Wps) GetGroups() (*GroupResult, error) {
	queryResult := GroupResult{}

	req, _ := http.NewRequest("GET", "https://www.kdocs.cn/3rd/drive/api/v5/groups/special/files", nil)
	req.Header.Add("Cookie", "wps_sid="+wps.sid)
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
		return &queryResult, err
	}
	result, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(result, &queryResult)

	return &queryResult, err
}
