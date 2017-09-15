package tools

func SafeGo(fn func()) {
	go fn()
}
