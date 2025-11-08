package task

import (
	"github.com/zyedidia/generic/list"
)

func RepositionLastElementToFirst[V any](l *list.List[V]) {
	fromNode, toNode := l.Back, l.Front
	l.Remove(fromNode)
	l.InsertBefore(fromNode, toNode)
}
