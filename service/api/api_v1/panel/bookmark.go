package panel

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golang.org/x/net/html"
)

// BookmarkImportRequest 书签导入请求结构
type BookmarkImportRequest struct {
	HtmlContent string            `json:"htmlContent" binding:"required_without=Bookmarks"`
	Bookmarks   []models.Bookmark `json:"Bookmarks" binding:"required_without=HtmlContent"`
}

// Bookmark 书签管理API
type Bookmark struct {
}

// AddMultiple 批量添加书签
func (a *Bookmark) AddMultiple(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	var req BookmarkImportRequest

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	var bookmarks []models.Bookmark
	var err error

	// 根据请求类型处理
	if req.HtmlContent != "" {
		// 处理HTML内容
		bookmarks, err = parseBookmarkHTML(req.HtmlContent, userInfo.ID)
		if err != nil {
			apiReturn.Error(c, fmt.Sprintf("解析HTML书签失败: %v", err))
			return
		}
	} else {
		// 处理已解析的书签数据
		bookmarks = req.Bookmarks
		// 为每个书签设置用户ID
		for i := range bookmarks {
			bookmarks[i].UserId = userInfo.ID
		}
	}

	// 检查URL唯一性并过滤重复的书签
	uniqueBookmarks := filterUniqueBookmarks(bookmarks, userInfo.ID)

	// 批量插入数据库
	if len(uniqueBookmarks) > 0 {
		if err := global.Db.Create(&uniqueBookmarks).Error; err != nil {
			apiReturn.Error(c, "添加书签失败")
			return
		}
	}

	apiReturn.SuccessData(c, map[string]interface{}{
		"count": len(uniqueBookmarks),
		"list":  uniqueBookmarks,
	})
}

// parseBookmarkHTML 解析浏览器导出的HTML书签文件
func parseBookmarkHTML(htmlContent string, userId uint) ([]models.Bookmark, error) {
	bookmarks := []models.Bookmark{}
	// 不再需要rootFolderID和folderMap，因为我们现在使用title作为URL和parentUrl

	// 解析HTML
	reader := bytes.NewReader([]byte(htmlContent))
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	// 递归解析HTML节点，第一层的parentUrl为"0"
	parseNode(doc, "0", &bookmarks, userId)

	return bookmarks, nil
}

// parseNode 递归解析HTML节点，按照用户要求处理层级关系
// parentUrl参数现在表示父级的URL，而不是ID
func parseNode(n *html.Node, parentUrl string, bookmarks *[]models.Bookmark, userId uint) {
	// 处理DL标签（文件夹容器）
	if n.Type == html.ElementNode && n.Data == "dl" {
		// 递归处理DL标签内的所有DT子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "dt" {
				// 只处理DT标签，确保正确的层级关系
				processDTNode(c, parentUrl, bookmarks, userId)
			}
		}
		return
	}

	// 根节点或其他容器节点的处理
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parseNode(c, parentUrl, bookmarks, userId)
	}
}

// processDTNode 专门处理DT标签，解析文件夹和书签
func processDTNode(n *html.Node, parentUrl string, bookmarks *[]models.Bookmark, userId uint) {
	// 检查是否包含H3标签（文件夹）
	hasH3 := false
	var folderName string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "h3" {
			hasH3 = true
			folderName = getNodeText(c)
			if folderName != "" {
				// 创建文件夹，URL设置为文件夹名称
				folder := models.Bookmark{
					Title:     folderName,
					Url:       folderName,
					IsFolder:  1,
					ParentUrl: parentUrl,
					UserId:    userId,
				}
				*bookmarks = append(*bookmarks, folder)
			}
			break
		}
	}

	// 检查当前DT节点内部是否包含DL标签（嵌套文件夹）
	if hasH3 && folderName != "" {
		foundNestedDL := false
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "dl" {
				// 递归处理嵌套的文件夹内容
				parseNode(c, folderName, bookmarks, userId)
				foundNestedDL = true
				break
			}
		}

		// 如果没有找到嵌套的DL标签，再查找下一个兄弟节点中的DL标签
		if !foundNestedDL {
			nextNode := n.NextSibling
			for nextNode != nil {
				if nextNode.Type == html.ElementNode && nextNode.Data == "dl" {
					// 递归处理该文件夹下的内容，使用文件夹名称作为父级URL
					parseNode(nextNode, folderName, bookmarks, userId)
					break
				}
				nextNode = nextNode.NextSibling
			}
		}
		return
	}

	// 检查是否包含A标签（书签项）
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "a" {
			// 这是一个书签
			url := ""
			for _, attr := range c.Attr {
				if attr.Key == "href" {
					url = attr.Val
					break
				}
			}
			title := getNodeText(c)
			if url != "" {
				bookmark := models.Bookmark{
					Title:     title,
					Url:       url,
					LanUrl:    "",
					Sort:      9999,
					IsFolder:  0,
					ParentUrl: parentUrl,
					UserId:    userId,
				}
				*bookmarks = append(*bookmarks, bookmark)
			}
			break
		}
	}
}

// getNodeText 获取节点的文本内容
func getNodeText(n *html.Node) string {
	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text.WriteString(c.Data)
		}
	}
	return strings.TrimSpace(text.String())
}

// filterUniqueBookmarks 过滤重复的书签，确保URL唯一性
func filterUniqueBookmarks(bookmarks []models.Bookmark, userId uint) []models.Bookmark {
	uniqueBookmarks := []models.Bookmark{}
	// 使用复合键（parentUrl + url）来确保在相同父级下不重复
	uniqueKeys := make(map[string]bool)

	// 首先获取用户已有的所有书签和文件夹，用于检查重复
	existingEntries := make(map[string]bool)
	var existingBookmarks []models.Bookmark
	global.Db.Where("user_id = ?", userId).Find(&existingBookmarks)
	for _, b := range existingBookmarks {
		// 创建复合键：parentUrl + url
		key := b.ParentUrl + ":" + b.Url
		existingEntries[key] = true
	}

	// 过滤重复的书签和文件夹
	for _, b := range bookmarks {
		// 创建复合键：parentUrl + url
		key := b.ParentUrl + ":" + b.Url

		// 检查是否已存在相同父级下的相同URL
		if !uniqueKeys[key] && !existingEntries[key] {
			uniqueKeys[key] = true
			uniqueBookmarks = append(uniqueBookmarks, b)
		}
	}

	return uniqueBookmarks
}

// Add 添加单个书签
func (a *Bookmark) Add(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	var req models.Bookmark

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 设置用户ID
	req.UserId = userInfo.ID

	// 插入数据库
	if err := global.Db.Create(&req).Error; err != nil {
		apiReturn.Error(c, "添加书签失败")
		return
	}

	apiReturn.SuccessData(c, req)
}

// 定义树形结构的书签节点
type BookmarkNode struct {
	models.Bookmark
	Children []*BookmarkNode `json:"children"`
}

// GetList 获取书签列表，并构建树形结构
func (a *Bookmark) GetList(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	var bookmarks []models.Bookmark

	// 查询当前用户的所有书签
	if err := global.Db.Where("user_id = ?", userInfo.ID).Find(&bookmarks).Error; err != nil {
		apiReturn.Error(c, "获取书签列表失败")
		return
	}

	// 构建树形结构
	tree := buildBookmarkTree(bookmarks)

	apiReturn.ListData(c, tree, int64(len(tree)))
}

// buildBookmarkTree 根据ParentUrl构建书签树
func buildBookmarkTree(bookmarks []models.Bookmark) []*BookmarkNode {
	// 创建节点映射表，用于快速查找父节点
	nodeMap := make(map[string]*BookmarkNode) // 使用string类型键，兼容各种类型的ID
	var rootNodes []*BookmarkNode

	// 首先创建所有节点
	for _, bookmark := range bookmarks {
		// 创建当前节点的副本
		currentBookmark := bookmark
		node := &BookmarkNode{
			Bookmark: currentBookmark,
			Children: []*BookmarkNode{},
		}
		// 使用ID的字符串形式作为节点的唯一标识
		nodeMap[strconv.Itoa(int(bookmark.ID))] = node
		// 同时也存储文件夹的标题，以便处理导入的HTML书签
		if bookmark.IsFolder == 1 {
			nodeMap[bookmark.Title] = node
		}
	}

	// 构建树形结构
	for _, bookmark := range bookmarks {
		node := nodeMap[strconv.Itoa(int(bookmark.ID))]
		parentUrl := bookmark.ParentUrl

		// 如果是根节点（ParentUrl为0），添加到根节点列表
		if parentUrl == "0" || parentUrl == "" || parentUrl == "null" {
			rootNodes = append(rootNodes, node)
		} else {
			// 查找父节点并添加当前节点到父节点的子节点列表
			// 尝试将ParentUrl作为ID查找
			parentNode, exists := nodeMap[parentUrl]
			if !exists {
				// 如果找不到，尝试将ParentUrl转换为整数后再查找
				if parentId, err := strconv.Atoi(parentUrl); err == nil {
					parentNode, exists = nodeMap[strconv.Itoa(parentId)]
				}
			}
			if exists {
				parentNode.Children = append(parentNode.Children, node)
			} else {
				// 如果找不到父节点，将当前节点作为根节点
				rootNodes = append(rootNodes, node)
			}
		}
	}

	return rootNodes
}

// Update 更新书签
func (a *Bookmark) Update(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	type UpdateReq struct {
		ID        uint   `json:"id" binding:"required"`
		Title     string `json:"title" binding:"required"`
		Url       string `json:"url" binding:"required"`
		LanUrl    string `json:"lanUrl"`
		ParentUrl string `json:"parentUrl"`
		Sort      int    `json:"sort"`
	}
	var req UpdateReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 检查书签是否存在且属于当前用户
	var bookmark models.Bookmark
	if err := global.Db.Where("id = ? AND user_id = ?", req.ID, userInfo.ID).First(&bookmark).Error; err != nil {
		apiReturn.Error(c, "书签不存在或无权修改")
		return
	}

	// 更新书签信息
	updateData := map[string]interface{}{
		"Title":     req.Title,
		"Url":       req.Url,
		"LanUrl":    req.LanUrl,
		"ParentUrl": req.ParentUrl,
		"Sort":      req.Sort,
		"UpdatedAt": time.Now(),
	}

	if err := global.Db.Model(&bookmark).Updates(updateData).Error; err != nil {
		apiReturn.Error(c, "更新书签失败")
		return
	}

	// 查询更新后的书签信息
	updatedBookmark := models.Bookmark{}
	global.Db.Where("id = ?", req.ID).First(&updatedBookmark)

	apiReturn.SuccessData(c, updatedBookmark)
}

// Deletes 删除书签
func (a *Bookmark) Deletes(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	type DeleteReq struct {
		Ids []int `json:"ids" binding:"required"`
	}
	var req DeleteReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	// 检查并删除用户的书签
	if err := global.Db.Where("user_id = ? AND id IN ?", userInfo.ID, req.Ids).Delete(&models.Bookmark{}).Error; err != nil {
		apiReturn.Error(c, "删除书签失败")
		return
	}

	apiReturn.Success(c)
}
