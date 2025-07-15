package data

import "github.com/jinzhu/copier"

type ProjectMenu struct {
	Id         int64
	Pid        int64 // 父id
	Title      string
	Icon       string
	Url        string
	FilePath   string
	Params     string
	Node       string
	Sort       int
	Status     int
	CreateBy   int64
	IsInner    int
	Values     string
	ShowSlider int
}

func (*ProjectMenu) TableName() string {
	return "ms_project_menu"
}

type ProjectMenuChild struct {
	ProjectMenu
	StatusText string
	InnerTest  string
	FullUrl    string
	Children   []*ProjectMenuChild
}

// CovertChild 菜单数据结构转换的入口地址，找出顶级菜单
// 作用： 前端可以直接使用树形结构渲染多级菜单
func CovertChild(pms []*ProjectMenu) []*ProjectMenuChild {
	var pmcs []*ProjectMenuChild
	copier.Copy(&pmcs, pms)
	for _, v := range pmcs {
		v.StatusText = getStatus(v.Status)
		v.InnerTest = getInnerText(v.IsInner)
		v.FullUrl = getFullUrl(v.Url, v.Params, v.Values)
	}
	var childPmcs []*ProjectMenuChild
	// 递归
	for _, v := range pmcs {
		if v.Pid == 0 {
			pmc := &ProjectMenuChild{}
			copier.Copy(pmc, v)
			childPmcs = append(childPmcs, pmc)
		}
	}
	toChild(childPmcs, pmcs)
	return childPmcs
}

func getFullUrl(url string, params string, values string) string {
	if (params != "" && values != "") || values != "" {
		return url + "/" + values
	}
	return url
}

func getInnerText(inner int) string {
	if inner == 0 {
		return "导航"
	}
	if inner == 1 {
		return "内页"
	}
	return ""
}

func getStatus(status int) string {
	if status == 0 {
		return "禁用"
	}
	if status == 1 {
		return "使用中"
	}
	return ""
}

// toChild 递归构建完整的菜单树形结构
func toChild(childPmcs []*ProjectMenuChild, pmcs []*ProjectMenuChild) {
	for _, pmc := range childPmcs {
		for _, pm := range pmcs {
			if pmc.Id == pm.Pid { // 如果找到子菜单
				child := &ProjectMenuChild{}
				copier.Copy(child, pm)
				pmc.Children = append(pmc.Children, child) // 添加到父菜单的子集合
			}
		}
		toChild(pmc.Children, pmcs) // 递归处理子菜单的子菜单
	}
}
