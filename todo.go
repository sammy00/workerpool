package workerpool

import "context"

//type ContextualizedAction
type todo struct {
	ctx    context.Context
	action Action
	done   func(err error)
}

//func (todo *todo) Execute() {
//	todo.done(todo.action.Execute(todo.ctx))
//}
