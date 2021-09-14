package tibbers

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

type nodeType uint8

const (
	static nodeType = iota // default
	root
)

type node struct {
	path     string
	indices  string
	nType    nodeType
	priority uint32
	children []*node
	handles  []Handle
}

// 自增priority 并且排序子节点 重新创建indices
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	// Adjust position (move to front)
	// 根据权重 重新排序children
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		// Swap node positions
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]
	}
	// 根据权重 重新生成indices
	if newPos != pos {
		n.indices = n.indices[:newPos] + // Unchanged prefix, might be empty
			n.indices[pos:pos+1] + // The index char we move
			n.indices[newPos:pos] + n.indices[pos+1:] // Rest without char at 'pos'
	}

	return newPos
}

// addRoute adds a node with the given handle to the path.
// Not concurrency-safe!
func (n *node) addRoute(path string, handles ...Handle) {
	fullPath := path
	n.priority++

	// Empty tree
	if n.path == "" && n.indices == "" {
		n.insertChild(path, fullPath, handles...)
		n.nType = root
		return
	}

walk:
	for {
		// 寻找最长匹配前缀
		i := longestCommonPrefix(path, n.path)

		// 公共长度小于当前节点长度，新建节点，拷贝当前节点
		// 当前节点赋值 children 包含当前
		if i < len(n.path) {
			child := node{
				path:     n.path[i:],
				nType:    static,
				indices:  n.indices,
				children: n.children,
				handles:  n.handles,
				priority: n.priority - 1,
			}

			n.children = []*node{&child}
			// 包含子节点第一个字母的字符串
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i]
			n.handles = nil
		}
		// 公共长度小于传值的节点 新建节点
		if i < len(path) {
			// 传进来的路径 去掉公共长度
			path = path[i:]
			idxc := path[0]
			// Check if a child with the next path byte exists
			for i, c := range []byte(n.indices) {
				if c == idxc {
					// 如果有子节点第一个字符相同，递归创建
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}
			// 如果没有 添加到最后
			n.indices += string([]byte{idxc})
			child := &node{}
			n.children = append(n.children, child)
			n.incrementChildPrio(len(n.indices) - 1)
			n = child
			n.insertChild(path, fullPath, handles...)
			return
		}

		// Otherwise add handle to current node
		if n.handles != nil {
			panic("a handle is already registered for path '" + fullPath + "'")
		}
		n.handles = handles
		return
	}
}

func (n *node) insertChild(path, fullPath string, handles ...Handle) {
	n.path = path
	n.handles = handles
}

func (n *node) getValue(path string) (handles []Handle) {
walk: // Outer loop for walking the tree
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				path = path[len(prefix):]

				// If this node does not have a wildcard (param or catchAll)
				// child, we can just look up the next child node and continue
				// to walk down the tree
				idxc := path[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						n = n.children[i]
						continue walk
					}
				}

			}
		} else if path == prefix {
			handles = n.handles
			return
		}
		return nil
	}
}
