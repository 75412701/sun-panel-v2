
<div align=center>

<img src="./doc/images/logo.png" width="100" height="100" />

# Sun-Panel-V2

一个基于[Sun-Panel](https://github.com/hslr-s/sun-panel)   修改的版本,增加了浏览器的导入书签的功能,使其主页和书签功能分开

Sun-Panel-V2 一个服务器、NAS导航面板、Homepage、浏览器首页、书签。

个人自用版本,后续会持续更新完善,如果你也喜欢建议点亮右上角星星避免后续迷路
</div>


## ✨ 功能特性

- **区分内外网链接**：自动区分内外网链接
- **私密模式**: 右下角切换私密模式,需要校验密码才能进入添加主页书签
- **便签功能**: 右上角便签功能,方便多端临时传输文本和文件
- **个性导航页**：支持导航页自定义设置样式。
- **书签管理**：支持导入浏览器书签到书签管理中
- **数据缓存**：避免多次请求数据
- **自适应**： pc端和移动端样式自适应




更新内容:
1.区分内外网链接：自动判断打开内外网链接
以前需要通过右下角切换访问模式,特别不方便,现在优化为自动的了

<img  src="https://img.meituan.net/portalweb/ba18a85e1401b1f6a9577f0ee064bc9b2836604.png"/>
<img src="https://nos.netease.com/ysf/d50bb118b2723964b2b837b112bf2c5e.png">
2.私密模式: 右下角切换模式改为私密模式,需要校验密码才能进入
加了这个功能,你的那些小网站就不怕被人看到了
<img  src="https://img.meituan.net/portalweb/d42ba6f468f6453c7ffd3eb23f7257483057468.png"/>



3.便签功能: 右上角便签功能,方便多端临时传输文本和文件
这个功能也是刚需,有时手机想给电脑同步点文本内容或者文件很不方便,尤其备用机微信QQ都没安装的情况下,愁死我了
<img  src="https://img.meituan.net/portalweb/63ffa42f90531e13552dc2e0bbd0bf97884958.png"/>



4.个性导航页：支持导航页自定义设置样式。
加了个自动获取网络壁纸的功能,避免审美疲劳
<img  src="https://img.meituan.net/portalweb/7dbca78be911a9d4e7c872639c69a148774045.png"/>



5.书签管理：支持导入浏览器书签到书签管理中
这个是老功能了,现在是参考谷歌浏览器的书签管理功能抄的,他的书签管理功能很好用啊,搞不懂网上其他家为何不抄非要自己搞个又丑又难用
<img  src="https://img.meituan.net/portalweb/d0e114e8eb6a6c41512a4e3fd0e70f73131042.png"/>



有人说找不到首页图标在哪里加,登录后切换私密模式后就可以添加了
<img  src="https://img.meituan.net/portalweb/afd97fd81a0209cc20ff5fff94098448593788.png"/>


6.移动端优化
功能都和pc端一样,只是移动端的界面样式美化了一下

<img  src="https://img.meituan.net/portalweb/304d7fd1d4dddc674235d0f9d36b3e73488502.png" width="400" height="700"/>
<img  src="https://img.meituan.net/portalweb/d8aabbf5742d9c574bc9901348da380e63582.png" width="400" height="700"/>


添加数据首页不显示?无需重新登录,点击右下角刷新按钮即可.因为数据是缓存在本地,不是随时请求数据,所以这样设计,主要是为了响应快
<img  src="https://nos.netease.com/ysf/d01ace9eb18cde000c0c22861079db84.png" />


## 部署
本项目支持 Docker 或其他基于 Docker 的平台部署。<br>
1.编写docker-compose.yml文件<br>
2.运行docker-compose up -d<br>
3.打开 域名/ip:3002<br><br><br>
账号:admin<br>
密码:123456
### docker

```yml
version: "3.2"

services:
  sun-panel:
    image: 'ghcr.io/75412701/sun-panel-v2:latest'
    container_name: sun-panel-v2
    volumes:
      - ./conf:/app/conf
      - ./uploads:/app/uploads
      - ./database:/app/database
    # - ./runtime:/app/runtime
    ports:
      - 3002:3002
    restart: always
```




## ❤️ Thanks

- [红烧猎人](https://blog.enianteam.com/u/sun/content/11)

---

[![Star History Chart](https://api.star-history.com/svg?repos=75412701/sun-panel-v2&type=Date)](https://star-history.com/#75412701/sun-panel-v2&Date)
