package shipping

var (
	_ Rule = (*WeightRule)(nil)
	_ Rule = (*ZoneRule)(nil)
	_ Rule = stubRule{}
)
