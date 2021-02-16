# wps-alfred-workflow

alfred workflow for WPS

## 登录

1. 登录网页版[wps](kdocs.cn), 登录成功后, 在 Cookie 中获取 wps_sid

2. `wps {wps_sid}`

![登录](.github/screen-shots/wps-login.png)

wps_sid 会保存在钥匙串中

## 查询最近文档

`wps`

![登录](.github/screen-shots/wps-recent.png)

## 使用关键词查询

`wps {keyword}`

![登录](.github/screen-shots/wps-search.png)