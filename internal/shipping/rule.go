package shipping

type LineItem struct {
	Rule string `json:"rule"`

	Description string `json:"description"`
	Amount      Money  `json:"amount"`
}

type Rule interface {
	Name() string

	Apply(Parcel) (LineItem, error)
}
