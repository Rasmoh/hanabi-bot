package main

var (
	strategies map[string]func() *Strategy
)

type Strategy struct {
	Name           string
	Start          func(*Strategy)
	GetAction      func(*Strategy, *Game) *Action
	ActionHappened func(*Strategy, *Game)
	Data           interface{}
}

func stratInit() {
	strategies = make(map[string]func() *Strategy)

	// Add a line below for each new strategy
	strategies["Dumb"] = NewDumb
	strategies["Hyphen-ated"] = NewHyphenated
}
