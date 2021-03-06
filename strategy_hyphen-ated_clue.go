package main

import (
	"strconv"
)

const (
	hyphenClueTypeSave = iota
	hyphenClueTypePlay
	// hyphenClueTypeFix
)

type PossibleClue struct {
	Clue       *Clue
	Target     int
	CardsClued int
}

/*
	Main functions
*/

func (d *Hyphenated) Check5Stall(g *Game) *Action {
	for i, p := range g.Players {
		if i == d.Us {
			continue
		}

		for _, c := range p.Hand {
			if c.Rank == 5 && len(c.Clues) == 0 {
				return &Action{
					Type:   ActionTypeRankClue,
					Target: p.Index,
					Value:  c.Rank,
				}
			}
		}
	}

	return nil
}

func (d *Hyphenated) Check5Burn(g *Game) *Action {
	for i, p := range g.Players {
		if i == d.Us {
			continue
		}

		for _, c := range p.Hand {
			if c.Rank == 5 {
				return &Action{
					Type:   ActionTypeRankClue,
					Target: p.Index,
					Value:  c.Rank,
				}
			}
		}
	}

	return nil
}

/*
	Subroutines
*/

func (d *Hyphenated) GetClueFocus(g *Game, i int, clue *Clue) *Card {
	p := g.Players[i]
	hp := d.Players[i]

	freshCards := p.GetFreshCardsTouchedByClue(g, clue)

	if len(freshCards) == 1 {
		// The focus of the clue is on the only brand new card introduced.
		return freshCards[0]
	}

	if len(freshCards) > 1 {
		// If one of the brand new cards introduced is on the chop, the focus is the chop.
		for _, c := range freshCards {
			if c == hp.GetChop(g, d) {
				return c
			}
		}

		// Otherwise, the focus is the left-most of the freshly touched cards.
		// Check to see if any of the freshly touched cards are in slot 1, then in slot 2,
		// and so forth.
		for i := 1; ; i++ {
			for _, c := range freshCards {
				if c.Slot == i {
					return c
				}
			}
		}
	}

	// If no brand new cards were introduced, the focus of the clue is the left-most touched card.
	touchedCards := p.GetCardsTouchedByClue(g, clue)
	for i := 1; i <= len(p.Hand); i++ {
		for _, c := range touchedCards {
			if c.Slot == i {
				return c
			}
		}
	}

	logger.Fatal("Failed to find the clue focus for the clue of: " + clue.Name(g))
	return nil
}

func (d *Hyphenated) CheckViableClue(g *Game, i int, j int, k int, clueType int) *PossibleClue {
	clue := &Clue{
		Type:  j,
		Value: k,
	}
	p := g.Players[i]
	hp := d.Players[i]
	touchedCards := p.GetCardsTouchedByClue(g, clue)

	// We are not allowed to give a clue that touches 0 cards in the hand.
	if len(touchedCards) == 0 {
		return nil
	}

	// Check if Good Touch Principle (1/2).
	// (e.g. if any of the touched cards are duplicates of one another)
	if len(touchedCards) >= 2 {
		for _, c := range touchedCards {
			for _, c2 := range touchedCards {
				if c == c2 {
					continue
				}
				if c.Suit == c2.Suit && c.Rank == c2.Rank {
					//logger.Debug("Clue " + clue.Name(g) + " failed because the touched cards contain a duplicate of each other.")
					return nil
				}
			}
		}
	}

	// Check for Good Touch Principle (2/2).
	// (e.g. if any of the touched cards are already touched in someone else's hand)
	freshCards := p.GetFreshCardsTouchedByClue(g, clue)
	for _, c := range freshCards {
		for i, p := range g.Players {
			for _, c2 := range p.Hand {
				if i == d.Us {
					// Don't potentially duplicate clued cards in our hand.
					mapIndex := c.Suit.Name + strconv.Itoa(c.Rank)
					if c2.Touched && c2.PossibleCards[mapIndex] > 0 {
						//logger.Debug("Clue " + clue.Name(g) + " failed because it could potentially duplicate a card in our hand.")
						return nil
					}
				} else {
					// Don't duplicate cards in other players hands.
					if c2.Touched && c.Suit == c2.Suit && c.Rank == c2.Rank {
						//logger.Debug("Clue " + clue.Name(g) + " failed because it would duplicate a card in another player's hand.")
						return nil
					}
				}
			}
		}
	}

	if clueType == hyphenClueTypePlay {
		c := d.GetClueFocus(g, i, clue)
		hc := d.Cards[c.Order] // nolint:staticcheck

		// Check to see if the card would misplay if we clued it.
		if c == nil || (!c.IsPlayable(g) && !hc.IsDelayedPlayable(g, d)) {
			//logger.Debug("Clue " + clue.Name(g) + " failed because the focus of the clue would misplay.")
			return nil
		}

		// Check to see if it will be interpreted as a 2 Save or a 5 Save.
		if c == hp.GetChop(g, d) && j == ClueTypeRank && (k == 2 || k == 5) {
			//logger.Debug("Clue " + clue.Name(g) + " failed because it would be interpreted as a 2 Save or a 5 Save.")
			return nil
		}
	}

	return &PossibleClue{
		Clue: &Clue{
			Type:  j,
			Value: k,
		},
		Target:     i,
		CardsClued: len(freshCards),
	}
}
