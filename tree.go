//go:build !rbtree
// +build !rbtree

package auction

import (
	avl "github.com/kjx98/go-avl"
)

type Tree struct {
	tree *avl.Tree[simOrderType]
}

type Iterator struct {
	tree *avl.Tree[simOrderType]
	it   *avl.Iterator[simOrderType]
}

//type TreeNode = avl.Node

func NewTree(cmpF avl.CompareFunc[simOrderType]) *Tree {
	var tree = Tree{}
	tree.tree = avl.New(cmpF)
	if tree.tree != nil {
		return &tree
	}
	// must no way go here
	return nil
}

func (t *Tree) destroy() {
	iter := t.tree.Iterator(avl.Forward)
	for node := iter.First(); node != nil; node = iter.Next() {
		t.tree.Remove(node)
	}
	t.tree = nil
}

func (t *Tree) Len() int {
	return t.tree.Len()
}

func (t *Tree) Find(key *simOrderType) *simOrderType {
	if node := t.tree.Find(key); node != nil {
		return &node.Value
	}
	return nil
}

func (t *Tree) Delete(key *simOrderType) bool {
	if v := t.tree.Find(key); v != nil {
		t.tree.Remove(v)
		return true
	}
	return false
}

func (t *Tree) Insert(v *simOrderType) {
	//or := v.(*simOrderType)
	//or.node.Value = v
	//t.tree.InsertNode(&or.node)
	t.tree.Insert(v)
}

func (t *Tree) First() *Iterator {
	it := Iterator{tree: t.tree}
	it.it = t.tree.Iterator(avl.Forward)
	it.it.First()
	return &it
}

func (it *Iterator) First() *simOrderType {
	if node := it.it.First(); node != nil {
		return &node.Value
	}
	return nil
}

func (it *Iterator) Get() *simOrderType {
	if node := it.it.Get(); node != nil {
		return &node.Value
	}
	return nil
}

func (it *Iterator) Next() *simOrderType {
	if node := it.it.Next(); node != nil {
		return &node.Value
	}
	return nil
}

func (it *Iterator) RemoveFirst() bool {
	if it.it == nil {
		return false
	}
	if node := it.it.Get(); node != nil {
		it.tree.Remove(node)
		// if it.it.Next() is nil, may set it.it to nil
		it.it.Next()
		return true
	}
	return false
}
