// +build rbtree

package auction

import (
	"github.com/kjx98/rbtree"
)

type Tree struct {
	tree *rbtree.Tree
}

type Iterator struct {
	tree *rbtree.Tree
	it   *rbtree.Iterator
}

func NewTree(cmpF rbtree.CompareFunc) *Tree {
	var tree = Tree{}
	tree.tree = rbtree.NewTree(cmpF)
	if tree.tree != nil {
		return &tree
	}
	// must no way go here
	return nil
}

func (t *Tree) destroy() {
	for iter := t.tree.Min(); !iter.Limit(); iter = t.tree.Min() {
		t.tree.DeleteWithKey(iter.Item())
	}
	t.tree = nil
}

func (t *Tree) Len() int {
	return t.tree.Len()
}

func (t *Tree) Find(key interface{}) interface{} {
	return t.tree.Get(key)
}

func (t *Tree) Delete(key interface{}) bool {
	if v := t.tree.Get(key); v != nil {
		t.tree.DeleteWithKey(v)
		return true
	}
	return false
}

func (t *Tree) Insert(v interface{}) {
	t.tree.Insert(v)
}

func (t *Tree) First() *Iterator {
	it := Iterator{tree: t.tree}
	it.it = t.tree.Min()
	return &it
}

func (it *Iterator) First() interface{} {
	it.it = it.tree.Min()
	return it.it.Item()
}

func (it *Iterator) Get() interface{} {
	if it.it.Limit() {
		return nil
	}
	return it.it.Item()
}

func (it *Iterator) Next() interface{} {
	if it.it.Limit() {
		return nil
	}
	it.it = it.it.Next()
	if it.it.Limit() {
		return nil
	}
	return it.it.Item()
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
