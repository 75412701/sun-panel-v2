package panel

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/siteFavicon"
	"sun-panel/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golang.org/x/net/html"
	"gorm.io/gorm"
)

// BookmarkImportRequest 书签导入请求结构
type BookmarkImportRequest struct {
	HtmlContent string            `json:"HtmlContent" binding:"required_without=Bookmarks"`
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

	// 按parentId分组处理书签的sort值
	// 1. 首先获取每个parentId下的当前最大sort值
	parentIdMaxSort := make(map[uint]int)

	// 收集所有需要处理的parentId
	parentIds := make(map[uint]bool)
	for _, b := range bookmarks {
		parentIds[b.ParentId] = true
	}

	// 为每个parentId获取当前最大sort值
	for parentId := range parentIds {
		var maxSort int
		query := global.Db.Model(&models.Bookmark{}).Where("user_id = ? AND parent_id = ?", userInfo.ID, parentId)
		query.Select("COALESCE(MAX(sort), 0) as max_sort").Scan(&maxSort)
		parentIdMaxSort[parentId] = maxSort
	}

	// 2. 为每个parentId下的书签设置递增的sort值
	// 先按parentId分组
	bookmarksByParentId := make(map[uint][]int) // key: parentId, value: 索引列表
	for i, b := range bookmarks {
		bookmarksByParentId[b.ParentId] = append(bookmarksByParentId[b.ParentId], i)
	}

	// 为每组书签设置递增的sort值
	for parentId, indexes := range bookmarksByParentId {
		currentMaxSort := parentIdMaxSort[parentId]
		for i, idx := range indexes {
			bookmarks[idx].Sort = currentMaxSort + i + 1
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

	// 处理新插入书签的parent_id关系（对于HTML导入的书签）
	// 1. 构建标题到ID的映射
	titleToIdMap := make(map[string]uint)
	for _, b := range uniqueBookmarks {
		if b.IsFolder == 1 {
			titleToIdMap[b.Title] = b.ID
		}
	}

	// 2. 更新子书签的parent_id
	for i, b := range uniqueBookmarks {
		if b.IsFolder == 0 && b.ParentUrl != "0" && b.ParentUrl != "" && b.ParentUrl != "null" {
			if parentId, exists := titleToIdMap[b.ParentUrl]; exists {
				uniqueBookmarks[i].ParentId = parentId
				// 更新数据库中的parent_id
				global.Db.Model(&models.Bookmark{}).Where("id = ? AND user_id = ?", b.ID, userInfo.ID).Update("parent_id", parentId)
			}
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
			iconJson := ""
			for _, attr := range c.Attr {
				if attr.Key == "href" {
					url = attr.Val
				} else if attr.Key == "icon" {
					iconJson = attr.Val
				}
			}
			// 检查是否有IMG子节点
			for imgChild := c.FirstChild; imgChild != nil; imgChild = imgChild.NextSibling {
				if imgChild.Type == html.ElementNode && imgChild.Data == "img" {
					for _, imgAttr := range imgChild.Attr {
						if imgAttr.Key == "src" {
							iconJson = imgAttr.Val
							break
						}
					}
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
					IconJson:  iconJson,
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

// getFaviconBase64 从URL获取favicon并转换为base64
func getFaviconBase64(urlStr string) (string, error) {
	// Check if it's an HTTP URL and add protocol if missing
	if !siteFavicon.IsHTTPURL(urlStr) {
		// Try adding https:// to URLs without protocol
		urlStr = "https://" + urlStr
	}

	// Parse the URL to get the host for Google's service
	urlInfo, parseErr := url.Parse(urlStr)
	var googleFaviconURL string
	if parseErr == nil {
		// Use the host part for Google's favicon service
		googleFaviconURL = fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s", urlInfo.Host)
	} else {
		// Fall back to using the original URL if parsing fails
		googleFaviconURL = fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s", urlStr)
	}

	// Create an HTTP client with a timeout to avoid hanging requests
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create a request with User-Agent header to avoid being blocked
	createRequest := func(urlStr string) (*http.Request, error) {
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return nil, err
		}
		// Add User-Agent to simulate browser request
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
		return req, nil
	}

	// First try: get favicon from the website itself
	faviconURL, err := siteFavicon.GetOneFaviconURL(urlStr)
	if err == nil {
		// Try downloading from the website's own favicon URL
		req, err := createRequest(faviconURL)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				data, err := io.ReadAll(resp.Body)
				if err == nil {
					// Get Content-Type from response header
					contentType := resp.Header.Get("Content-Type")
					if contentType == "" {
						contentType = "image/png" // Default to PNG if not specified
					}
					return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
				}
			}
		}
		// If any step fails, fall through to try Google's service
	}

	// Second try: use Google's favicon service
	req, err := createRequest(googleFaviconURL)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	// Read the image data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Get Content-Type from response header
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png" // Default to PNG if not specified
	}

	// Encode to base64 and add data URL prefix
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
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

	// 获取parentId下的最大sort值并+1作为新的sort值
	var maxSort int
	// 使用COALESCE确保如果没有记录时返回0
	query := global.Db.Model(&models.Bookmark{}).Where("user_id = ? AND parent_id = ?", userInfo.ID, req.ParentId)
	query.Select("COALESCE(MAX(sort), 0) as max_sort").Scan(&maxSort)

	// 设置新的sort值
	req.Sort = maxSort + 1

	// 插入数据库（不包含图标数据）
	if err := global.Db.Create(&req).Error; err != nil {
		apiReturn.Error(c, "添加书签失败")
		return
	}

	// 返回结果给前端
	apiReturn.SuccessData(c, req)

	// 使用协程异步获取并更新图标
	if req.IsFolder == 0 {
		go func(id uint, url string, userId uint) {
			if faviconBase64, err := getFaviconBase64(url); err == nil && faviconBase64 != "" {
				// 更新书签的图标数据
				updateData := map[string]interface{}{
					"IconJson": faviconBase64,
				}
				global.Db.Model(&models.Bookmark{}).Where("id = ? AND user_id = ?", id, userId).UpdateColumns(updateData)
			}
		}(req.ID, req.Url, userInfo.ID)
	}
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

	// 优化查询排序：先按ParentUrl分组，再按sort字段升序排序，确保同一父文件夹下的项目按正确顺序返回
	// 这样可以保证在构建树形结构时，子节点的添加顺序就是排序后的顺序
	if err := global.Db.Where("user_id = ?", userInfo.ID).Order("sort ASC").Find(&bookmarks).Error; err != nil {
		apiReturn.Error(c, "获取书签列表失败")
		return
	}

	// 构建树形结构
	tree := buildBookmarkTree(bookmarks)

	apiReturn.ListData(c, tree, int64(len(tree)))
}

// sortNodesBySortField 根据sort字段对节点数组进行升序排序
func sortNodesBySortField(nodes []*BookmarkNode) {
	// 按sort字段升序排序，如果sort相同则按标题排序
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Sort != nodes[j].Sort {
			return nodes[i].Sort < nodes[j].Sort
		}
		return nodes[i].Title < nodes[j].Title
	})
}

// sortChildNodes 递归对节点的子节点进行排序
func sortChildNodes(node *BookmarkNode) {
	if len(node.Children) > 0 {
		// 对子节点排序
		sortNodesBySortField(node.Children)
		// 递归处理每个子节点
		for _, child := range node.Children {
			sortChildNodes(child)
		}
	}
}

// buildBookmarkTree 根据ParentId构建书签树
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
	}

	// 构建树形结构
	for _, bookmark := range bookmarks {
		node := nodeMap[strconv.Itoa(int(bookmark.ID))]
		parentId := bookmark.ParentId

		// 如果是根节点（ParentId为0或空），添加到根节点列表
		if parentId == 0 {
			rootNodes = append(rootNodes, node)
		} else {
			// 查找父节点并添加当前节点到父节点的子节点列表
			if parentNode, exists := nodeMap[strconv.Itoa(int(parentId))]; exists {
				parentNode.Children = append(parentNode.Children, node)
			} else {
				// 如果找不到父节点，添加到根节点列表
				rootNodes = append(rootNodes, node)
			}
		}
	}

	// 确保对所有层级进行排序
	// 1. 首先对根节点排序
	sortNodesBySortField(rootNodes)

	// 2. 递归对所有子节点进行排序，确保每个文件夹下的子节点都按sort升序排列
	for _, rootNode := range rootNodes {
		sortChildNodes(rootNode)
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
		ParentId  uint   `json:"parentId"`
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

	// 准备更新数据
	updateData := map[string]interface{}{
		"Title":     req.Title,
		"Url":       req.Url,
		"LanUrl":    req.LanUrl,
		"ParentUrl": req.ParentUrl,
		"ParentId":  req.ParentId,
		// "UpdatedAt": time.Now(), // 移除UpdatedAt更新
	}

	// 如果parentId改变了，需要重新计算sort值
	if bookmark.ParentId != req.ParentId {
		// 获取新parentId下的最大sort值并+1
		var maxSort int
		query := global.Db.Model(&models.Bookmark{}).Where("user_id = ? AND parent_id = ?", userInfo.ID, req.ParentId)
		query.Select("COALESCE(MAX(sort), 0) as max_sort").Scan(&maxSort)
		updateData["Sort"] = maxSort + 1
	} else {
		// 如果parentId没变，保留原有的sort逻辑
		updateData["Sort"] = req.Sort
	}

	// 保存更新（不包含图标数据）
	if err := global.Db.Model(&models.Bookmark{}).Where("id = ? AND user_id = ?", bookmark.ID, userInfo.ID).Updates(updateData).Error; err != nil {
		apiReturn.Error(c, "修改书签失败")
		return
	}

	// 查询更新后的书签信息
	updatedBookmark := models.Bookmark{}
	if err := global.Db.Where("id = ? AND user_id = ?", req.ID, userInfo.ID).First(&updatedBookmark).Error; err != nil {
		// 如果查询失败，返回更新前的基本信息
		updatedBookmark := bookmark
		updatedBookmark.Title = req.Title
		updatedBookmark.Url = req.Url
		updatedBookmark.LanUrl = req.LanUrl
		updatedBookmark.ParentUrl = req.ParentUrl
		updatedBookmark.ParentId = req.ParentId
		updatedBookmark.Sort = updateData["Sort"].(int)
		apiReturn.SuccessData(c, updatedBookmark)
	} else {
		// 返回更新后的信息
		apiReturn.SuccessData(c, updatedBookmark)
	}

	// 如果URL改变且是书签（非文件夹），异步重新获取favicon
	if req.Url != bookmark.Url && bookmark.IsFolder == 0 {
		go func(id uint, url string, userId uint) {
			if faviconBase64, err := getFaviconBase64(url); err == nil && faviconBase64 != "" {
				// 更新书签的图标数据
				updateIconData := map[string]interface{}{
					"IconJson": faviconBase64,
				}
				global.Db.Model(&models.Bookmark{}).Where("id = ? AND user_id = ?", id, userId).UpdateColumns(updateIconData)
			}
		}(bookmark.ID, req.Url, userInfo.ID)
	}
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

	// 收集所有需要删除的书签ID，包括文件夹下的所有子项
	var allIdsToDelete []int

	// 遍历每个要删除的ID
	for _, id := range req.Ids {
		// 首先检查该ID是否属于当前用户
		var bookmark models.Bookmark
		if err := global.Db.Where("user_id = ? AND id = ?", userInfo.ID, id).First(&bookmark).Error; err != nil {
			continue // 跳过不属于当前用户的书签
		}

		// 添加到要删除的列表中
		allIdsToDelete = append(allIdsToDelete, id)

		// 如果是文件夹，需要删除其下所有的书签和子文件夹
		if bookmark.IsFolder == 1 {
			// 递归获取所有子项ID，使用文件夹的ID作为parent_id
			childIds := getAllChildBookmarkIds(userInfo.ID, bookmark.ID)
			allIdsToDelete = append(allIdsToDelete, childIds...)
		}
	}

	// 如果没有有效的ID需要删除，直接返回成功
	if len(allIdsToDelete) == 0 {
		apiReturn.Success(c)
		return
	}

	// 删除所有收集到的书签，使用Unscoped()确保是硬删除
	if err := global.Db.Unscoped().Where("user_id = ? AND id IN ?", userInfo.ID, allIdsToDelete).Delete(&models.Bookmark{}).Error; err != nil {
		apiReturn.Error(c, "删除书签失败")
		return
	}

	apiReturn.Success(c)
}

// BookmarkSortItem 单个书签排序项
type BookmarkSortItem struct {
	Id       uint  `json:"id" binding:"required"`
	Sort     int   `json:"sort"`
	ParentId *uint `json:"parentId"`
}

// BookmarkSaveSortRequest 批量保存书签排序请求
type BookmarkSaveSortRequest struct {
	SortItems []BookmarkSortItem `json:"sortItems" binding:"required"`
}

// SaveSort 批量保存书签排序
func (a *Bookmark) SaveSort(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	var req BookmarkSaveSortRequest

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	err := global.Db.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.SortItems {
			updateData := map[string]interface{}{
				"Sort": item.Sort,
			}
			if item.ParentId != nil {
				updateData["ParentId"] = *item.ParentId
			}
			if err := tx.Model(&models.Bookmark{}).
				Where("user_id = ? AND id = ?", userInfo.ID, item.Id).
				Updates(updateData).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.Success(c)
}

// getAllChildBookmarkIds 递归获取文件夹下所有子书签和子文件夹的ID
func getAllChildBookmarkIds(userId uint, parentId uint) []int {
	var childIds []int
	var childBookmarks []models.Bookmark

	// 查询直接子项
	if err := global.Db.Where("user_id = ? AND parent_id = ?", userId, parentId).Find(&childBookmarks).Error; err != nil {
		return childIds
	}

	for _, child := range childBookmarks {
		childIds = append(childIds, int(child.ID))
		// 如果子项也是文件夹，递归获取其子项
		if child.IsFolder == 1 {
			grandChildIds := getAllChildBookmarkIds(userId, child.ID)
			childIds = append(childIds, grandChildIds...)
		}
	}

	return childIds
}
