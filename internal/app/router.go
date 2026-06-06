package app

import "gitflow-tui/internal/tui"

// Router manages screen navigation.
type Router struct {
	current tui.Screen
}

// NewRouter creates a new router.
func NewRouter(initial tui.Screen) *Router {
	return &Router{current: initial}
}

// Current returns the current screen.
func (r *Router) Current() tui.Screen {
	return r.current
}

// Navigate switches to a new screen.
func (r *Router) Navigate(screen tui.Screen) {
	r.current = screen
}
