package data

import "strings"

type ProjectNode struct {
	Id       int64
	Node     string
	Title    string
	IsMenu   int
	IsLogin  int
	IsAuth   int
	CreateAt int64
}

func (*ProjectNode) TableName() string {
	return "ms_project_node"
}

type ProjectNodeTree struct {
	Id       int64
	Node     string
	Title    string
	IsMenu   int
	IsLogin  int
	IsAuth   int
	Pnode    string
	Children []*ProjectNodeTree
}

// node节点 project project/index
// 分而治之：先处理根节点，再递归处理子节点
// 路径解析：通过 "/" 分割符确定节点层级关系
// 递归构建：自顶向下逐层构建完整树形结构

func ToNodeTreeList(list []*ProjectNode) []*ProjectNodeTree {
	var roots []*ProjectNodeTree
	for _, v := range list {
		paths := strings.Split(v.Node, "/")
		if len(paths) == 1 {
			// 说明是根节点
			root := &ProjectNodeTree{
				Id:       v.Id,
				Node:     v.Node,
				Pnode:    "",
				IsLogin:  v.IsLogin,
				IsMenu:   v.IsMenu,
				IsAuth:   v.IsAuth,
				Title:    v.Title,
				Children: []*ProjectNodeTree{},
			}
			roots = append(roots, root)
		}
	}
	for _, v := range roots {
		addChild(list, v, 2)
	}
	return roots
}

func addChild(list []*ProjectNode, root *ProjectNodeTree, level int) {
	for _, v := range list {
		paths := strings.Split(v.Node, "/")
		if len(paths) == level && strings.HasPrefix(v.Node, root.Node+"/") {
			child := &ProjectNodeTree{
				Id:       v.Id,
				Node:     v.Node,
				Pnode:    "",
				IsLogin:  v.IsLogin,
				IsMenu:   v.IsMenu,
				IsAuth:   v.IsAuth,
				Title:    v.Title,
				Children: []*ProjectNodeTree{},
			}
			root.Children = append(root.Children, child)
		}
	}
	for _, v := range root.Children {
		addChild(list, v, level+1)
	}
}

type ProjectNodeAuthTree struct {
	Id       int64
	Node     string
	Title    string
	IsMenu   int
	IsLogin  int
	IsAuth   int
	Pnode    string
	Key      string
	Checked  bool
	Children []*ProjectNodeAuthTree
}

func ToAuthNodeTreeList(list []*ProjectNode, checkedList []string) []*ProjectNodeAuthTree {
	checkedMap := make(map[string]struct{})
	for _, v := range checkedList {
		checkedMap[v] = struct{}{} // 空结构体类型的零值实例，优势：每个值占用 0 字节内存，这样的操作实现了set数据结构
	}
	var roots []*ProjectNodeAuthTree
	for _, v := range list {
		paths := strings.Split(v.Node, "/")
		if len(paths) == 1 {
			// 检查该节点是否已授权
			checked := false
			if _, ok := checkedMap[v.Node]; ok {
				checked = true
			}
			//根节点
			root := &ProjectNodeAuthTree{
				Id:       v.Id,
				Node:     v.Node,
				Pnode:    "",
				IsLogin:  v.IsLogin,
				IsMenu:   v.IsMenu,
				IsAuth:   v.IsAuth,
				Title:    v.Title,
				Children: []*ProjectNodeAuthTree{},
				Checked:  checked, // 权限状态
				Key:      v.Node,  // 用于前端组件的 key
			}
			roots = append(roots, root)
		}
	}
	for _, v := range roots {
		addAuthNodeChild(list, v, 2, checkedMap)
	}
	return roots
}

func addAuthNodeChild(list []*ProjectNode, root *ProjectNodeAuthTree, level int, checkedMap map[string]struct{}) {
	for _, v := range list {
		if strings.HasPrefix(v.Node, root.Node+"/") && len(strings.Split(v.Node, "/")) == level {
			// 此根节点子节点
			// 权限检查
			checked := false
			if _, ok := checkedMap[v.Node]; ok {
				checked = true
			}

			child := &ProjectNodeAuthTree{
				Id:       v.Id,
				Node:     v.Node,
				Pnode:    "",
				IsLogin:  v.IsLogin,
				IsMenu:   v.IsMenu,
				IsAuth:   v.IsAuth,
				Title:    v.Title,
				Children: []*ProjectNodeAuthTree{},
				Checked:  checked,
				Key:      v.Node,
			}
			root.Children = append(root.Children, child)
		}
	}
	for _, v := range root.Children {
		addAuthNodeChild(list, v, level+1, checkedMap)
	}
}
