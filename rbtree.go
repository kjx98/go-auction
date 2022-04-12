//go:build rbtree
// +build rbtree

package auction

import (
	"github.com/kjx98/rbtree"
)

type Tree struct {
	tree *rbtree.Tree[*simOrderType]
}

type Iterator struct {
	tree *rbtree.Tree[*simOrderType]
	it   *rbtree.Iterator[*simOrderType]
}

//type TreeNode = rbtree.Node

func NewTree(cmpF rbtree.CompareFunc[*simOrderType]) *Tree {
	var tree = Tree{}
	tree.tree = rbtree.New[*simOrderType](cmpF)
	if tree.tree != nil {
		return &tree
	}
	// must no way go here
	return nil
}

func (t *Tree) destroy() {
	for iter := t.tree.Min(); !iter.Limit(); iter = t.tree.Min() {
		t.tree.DeleteWithKey(*iter.Item())
	}
	t.tree = nil
}

func (t *Tree) Len() int {
	return t.tree.Len()
}

func (t *Tree) Find(key *simOrderType) *simOrderType {
	return *t.tree.Find(key)
}

func (t *Tree) Delete(key *simOrderType) bool {
	return t.tree.DeleteWithKey(key)
}

func (t *Tree) Insert(v *simOrderType) {
	t.tree.Insert(v)
}

func (t *Tree) First() *Iterator {
	it := Iterator{tree: t.tree}
	it.it = t.tree.Min()
	return &it
}

func (it *Iterator) First() *simOrderType {
	it.it = it.tree.Min()
	if it.it.Limit() {
		return nil
	}
	return *it.it.Item()
}

func (it *Iterator) Get() *simOrderType {
	if it.it.Limit() {
		return nil
	}
	return *it.it.Item()
}

func (it *Iterator) Next() *simOrderType {
	if it.it.Limit() {
		return nil
	}
	it.it = it.it.Next()
	if it.it.Limit() {
		return nil
	}
	return *it.it.Item()
}

func (it *Iterator) RemoveFirst() bool {
	// current node must be first
	if it.it.Limit() {
		return false
	}
	itNext := it.it.Next()
	it.tree.DeleteWithIterator(it.it)
	it.it = itNext
	return true
}
