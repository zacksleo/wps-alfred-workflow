package main

import (
	"fmt"
	"os"
	"strings"

	aw "github.com/deanishe/awgo"
)

var (
	recentCacheName = "recent.json"
	groupCacheName  = "group.json"
)

type User struct {
	wf          *aw.Workflow
	Wps         *Wps
	wpsCacheDir string
}

// GroupFile struct
type GroupFile struct {
	GroupID     string `json:"groupid"`
	IsLinkGroup bool   `json:"islinkgroup"`
	FileID      int64  `json:"fileid"`
}

func getGroupFile(wf *aw.Workflow, text string) GroupFile {
	groupFile := GroupFile{}
	wf.Cache.LoadJSON(getMd5(text)+".json", &groupFile)
	return groupFile
}

//查询最近的文档
func (user *User) getLatest() {
	wf := user.wf

	files := []LatestFile{}

	if wf.Cache.Exists(recentCacheName) {
		wf.Cache.LoadJSON(recentCacheName, &files)
	}

	if wf.Cache.Expired(recentCacheName, getCacheAge("latest_expire_mins")) {
		result, err := user.Wps.GetLatest()
		if err != nil {
			wf.WarnEmpty("查询失败", err.Error())
			return
		}
		files = *result
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
			quickLookURL = user.wpsCacheDir + user.getFilePath(file.GroupID, file.FileID)
		}
		item := wf.NewItem(file.Name).
			Subtitle(fmt.Sprintf("%s %s上阅读 %s", getTimeDiff(file.Mtime/1000), file.OriginalDeviceType, file.OriginalDeviceName)).
			Valid(true).Icon(getIcon(file.Name)).
			Var("fileid", file.FileID).Var("name", file.Name).Var("dir", quickLookURL).
			Var("url", fmt.Sprintf("https://www.kdocs.cn/p/%s", file.FileID)).
			Quicklook(quickLookURL)
		item.Opt().Subtitle("复制分享连接")
		item.Cmd().Subtitle("在 WPS 中查看")
		item.Ctrl().Subtitle("在 Finder 中查看")
	}
	wf.SendFeedback()
}

func (user *User) queryDocs(query string) {
	wf := user.wf
	queryResult := QueryResult{}
	cacheName := query + ".json"

	if wf.Cache.Exists(cacheName) {
		wf.Cache.LoadJSON(cacheName, &queryResult.Files)
	}

	if wf.Cache.Expired(cacheName, getCacheAge("query_expire_mins")) {
		result, err := user.Wps.QueryDocs(query)

		if err != nil {
			wf.WarnEmpty("查询失败", err.Error())
			return
		}
		queryResult = *result

		wf.Cache.StoreJSON(cacheName, queryResult.Files)
	}

	if len(queryResult.Files) == 0 {
		wf.WarnEmpty("没有找到文件", "请使用其他关键词搜索")
	}

	for _, file := range queryResult.Files {
		quickLookURL := user.wpsCacheDir + "/" + strings.Replace(file.Path, "我的云文档", "团队文档", 1) + "/" + file.Fname
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
func (user *User) getGroups() {
	wf := user.wf
	queryResult := GroupResult{}

	if wf.Cache.Exists(groupCacheName) {
		wf.Cache.LoadJSON(groupCacheName, &queryResult.Files)
	}

	if wf.Cache.Expired(groupCacheName, getCacheAge("groups_expire_mins")) {

		result, err := user.Wps.GetGroups()

		if err != nil {
			wf.WarnEmpty("查询失败", err.Error())
			return
		}

		queryResult = *result

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
		item := wf.NewItem(file.Fname).
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

func (user *User) getGroupFiles(path string) {
	wf := user.wf
	groupID := os.Getenv("groupid")
	parentID := os.Getenv("fileid")
	parentFileID := os.Getenv("parentFileid")
	isLinkGroup := true
	if groupID == "" {
		groupFile := getGroupFile(wf, path)
		groupID = groupFile.GroupID
		isLinkGroup = groupFile.IsLinkGroup
		parentID = fmt.Sprintf("%d", groupFile.FileID)
	}

	queryResult := FileResult{}
	cacheName := fmt.Sprintf("%s-%s.json", groupID, parentID)

	if wf.Cache.Exists(cacheName) {
		wf.Cache.LoadJSON(cacheName, &queryResult.Files)
	}

	if wf.Cache.Expired(cacheName, getCacheAge("group_file_expire_mins")) {

		result, err := user.Wps.GetGroupFiles(groupID, parentID)

		if err != nil {
			wf.WarnEmpty("查询失败", err.Error())
			return
		}
		queryResult = *result

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
		quickLookURL := user.wpsCacheDir + path + "/" + file.Fname
		url := fmt.Sprintf("https://www.kdocs.cn/p/%d", file.ID)
		if isLinkGroup {
			quickLookURL = user.wpsCacheDir + "/团队文档" + path + "/" + file.Fname
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
			Var("groupid", groupID).Var("fileid", fmt.Sprintf("%d", file.ID)).Var("dir", quickLookURL).
			Var("parentFileid", parentID).
			Var("name", file.Fname).
			Var("url", url).
			Var("path", path+"/"+file.Fname).Valid(true).Autocomplete(path + "/" + file.Fname).
			Quicklook(quickLookURL)
		item.Opt().Subtitle("复制分享连接")
		item.Cmd().Subtitle("在 WPS 中查看")
		item.Ctrl().Subtitle("在 Finder 中查看")
	}
	wf.SendFeedback()
}

// 根据文件ID 和 GroupID 查询文件路径
func (user *User) getFilePath(groupID int, fileID string) string {
	wf := user.wf
	groupFilePathResult := GroupFilePathResult{}

	cacheName := fileID + "-path.json"
	if wf.Cache.Exists(cacheName) {
		wf.Cache.LoadJSON(cacheName, &groupFilePathResult.Path)
	}

	if wf.Cache.Expired(cacheName, getCacheAge("file_path_expire_mins")) {
		result, err := user.Wps.GetFilePath(groupID, fileID)
		if err != nil {
			wf.WarnEmpty("查询失败", err.Error())
			return ""
		}

		groupFilePathResult = *result

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
