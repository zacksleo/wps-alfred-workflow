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
	"log"
	"strings"

	aw "github.com/deanishe/awgo"
)

var (
	wf *aw.Workflow
)

//注销登录时，删除钥匙串中存储的信息，删除缓存信息
func logout() {
	wf.Keychain.Delete("wps_sid")
	wf.ClearCache()
	wf.WarnEmpty("注销成功", "")
	wf.SendFeedback()
}

// 登录成功后，设置钥匙串
func login(user *User, query string) {
	if query == "" {
		wf.WarnEmpty("账号没有登录", "请先登录: wps {wps_sid}")
		wf.SendFeedback()
		return
	}
	wf.Keychain.Set("wps_sid", query)
	log.Printf("set keychain =%s\n", query)
	user.wpsSid = query
	user.getLatest()
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

	user := User{wf, wpsSid}

	// 如果没有找到钥匙串，则提醒用户登录
	if err != nil {
		login(&user, query)
		return
	}

	// 默认搜索最近的文档
	if len(query) < 1 {
		user.getLatest()
		return
	}

	if query == "/" {
		user.getGroups()
		return
	}

	if strings.HasPrefix(query, "/") {
		user.getGroupFiles(query)
		return
	}

	//使用关键词搜索
	user.queryDocs(query)
}

func main() {
	wf.Run(run)
}
