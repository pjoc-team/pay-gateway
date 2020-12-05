package manager

// NotInitializedError 未初始化错误
type NotInitializedError struct {
	Message string
}

func (n *NotInitializedError) Error() string {
	return n.Message
}

