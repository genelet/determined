package utils

import (
	"sync"

	ilang "github.com/genelet/determined/internal/lang"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

const (
	VAR        = "var"
	ATTRIBUTES = "attributes"
	FUNCTIONS  = "functions"
)

// Tree is a tree structure that can be used to represent a nested structure in HCL
type Tree struct {
	mu    sync.Mutex
	Name  string
	Data  sync.Map
	Up    *Tree
	Downs []*Tree
}

// NewTree creates a new tree with the given name
func NewTree(name string) *Tree {
	return &Tree{Name: name, Data: sync.Map{}}
}

// DefaultTreeFunctions returns the default tree, and the map containing default functions and the default tree
func DefaultTreeFunctions(ref map[string]interface{}) (*Tree, map[string]interface{}) {
	if ref == nil {
		ref = make(map[string]interface{})
	}
	var node *Tree
	if inode, ok := ref[ATTRIBUTES]; ok {
		node = inode.(*Tree)
	} else {
		node = NewTree(VAR)
		ref[ATTRIBUTES] = node
	}
	defaultFuncs := ilang.CoreFunctions(".")
	if ref[FUNCTIONS] == nil {
		ref[FUNCTIONS] = defaultFuncs
	} else if t, ok := ref[FUNCTIONS].(map[string]function.Function); ok {
		for k, v := range defaultFuncs {
			t[k] = v
		}
	}

	return node, ref
}

// AddNode adds a new node to the tree. It returns the newly created node.
func (self *Tree) AddNode(name string) *Tree {
	first := grep(self.Downs, name)
	if first != nil {
		return first
	}

	self.mu.Lock()
	child := NewTree(name)
	self.Downs = append(self.Downs, child)
	child.Up = self
	self.mu.Unlock()
	return child
}

// AddNodes adds a new node to the tree, and then adds the given names as children to the new node.
func (self *Tree) AddNodes(tag string, names ...string) *Tree {
	node := self.AddNode(tag)
	for _, name := range names {
		node = node.AddNode(name)
	}
	return node
}

func grep(downs []*Tree, name string) *Tree {
	for _, down := range downs {
		if down.Name == name {
			return down
		}
	}
	return nil
}

// GetNode returns the node with the given name. If the node does not exist, it returns nil.
func (self *Tree) GetNode(tag string, names ...string) *Tree {
	if tag == "" {
		return self
	}
	down := grep(self.Downs, tag)
	if down == nil {
		return nil
	}

	for _, name := range names {
		down = grep(down.Downs, name)
		if down == nil {
			return nil
		}
	}

	return down
}

// DeleteNode deletes the node with the given name.
func (self *Tree) DeleteNode(name string) {
	self.mu.Lock()
	for i, item := range self.Downs {
		if item.Name == name {
			if i+1 == len(self.Downs) {
				self.Downs = self.Downs[:i]
			} else {
				self.Downs = append(self.Downs[:i], self.Downs[i+1:]...)
			}
			self.mu.Unlock()
			return
		}
	}
	self.mu.Unlock()
}

// AddItem adds a new item the tree data.
func (self *Tree) AddItem(k string, v cty.Value) {
	self.Data.Store(k, v)
}

// GetItem returns the item with the given key in the tree data
func (self *Tree) DeleteItem(k string) {
	self.Data.Delete(k)
}

// GetItem returns the item with the given key.
func (self *Tree) FindNode(names []string) *Tree {
	if names == nil || len(names) == 0 {
		return nil
	}

	var down *Tree

	for _, item := range self.Downs {
		if item.Name == names[0] {
			if len(names) == 1 {
				return item
			}
			return item.FindNode(names[1:])
		} else {
			down = item.FindNode(names)
			if down != nil {
				return down
			}
		}
	}

	return nil
}

// Variables returns the variables which can be used as the HCL Value context.
func (self *Tree) Variables() map[string]cty.Value {
	hash := make(map[string]cty.Value)
	for _, down := range self.Downs {
		if variables := down.Variables(); variables != nil {
			hash[down.Name] = cty.ObjectVal(variables)
		}
	}
	self.Data.Range(func(k, v any) bool {
		if k != VAR {
			hash[k.(string)] = v.(cty.Value)
		}
		return true
	})

	if self.Name == VAR {
		hash[VAR] = cty.ObjectVal(hash)
	}

	return hash
}
