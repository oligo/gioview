package view

type defaultViewProvider struct {
	id      ViewID
	builder func() View
}

func (p defaultViewProvider) Provide(intent Intent) View {
	if intent.Target != p.id {
		panic("view routing error ")
	}

	if p.builder == nil {
		panic("view builder is missing")
	}

	return p.builder()

}

func (p defaultViewProvider) ID() ViewID {
	return p.id
}

// Provide is a helper function.
func Provide(viewID ViewID, builder func() View) ViewProvider {
	return defaultViewProvider{id: viewID, builder: builder}
}
