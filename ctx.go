package extensions

import "fmt"

type ctxKey string

func getCtxKey(key string, args []interface{}) ctxKey {
	if args == nil {
		return ctxKey(key)
	}
	for _, arg := range args {
		key += fmt.Sprintf("_%v", arg)
	}
	return ctxKey(key)
}

func getParentCtxKey(key string, args []interface{}) ctxKey {
	if len(args) <= 1 {
		return ctxKey(key)
	}
	if _, ok := args[len(args)-2].(int); ok {
		for i := 0; i < len(args)-2; i++ {
			key += fmt.Sprintf("_%v", args[i])
		}
		return ctxKey(key)
	}
	for i := 0; i < len(args)-1; i++ {
		key += fmt.Sprintf("_%v", args[i])
	}
	return ctxKey(key)
}
