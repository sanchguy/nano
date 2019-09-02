package game

import (
	"strconv"

	"github.com/sanchguy/nano"
)

/*
* Matrix used to calculate the card weight in the Truco games
*   weigth -  card
*      13  -  1 espada
*      12  -  1 basto
*      11  -  7 espada
*      10  -  7 oro
*       9  -  3
*       8  -  2
*       7  -  1 copa
*       7  -  1 oro
*       6  -  12
*       5  -  11
*       4  -  7 copa
*       4  -  7 basto
*       3  -  6
*       2  -  5
*       1  -  4
 */

var weights = map[string][]int{
	"gold":    {0, 8, 9, 10, 1, 2, 3, 11, 0, 0, 5, 6, 7},
	"cup":   {0, 8, 9, 10, 1, 2, 3, 4, 0, 0, 5, 6, 7},
	"sword": {0, 14, 9, 10, 1, 2, 3, 12, 0, 0, 5, 6, 7},
	"club":  {0, 13, 9, 10, 1, 2, 3, 4, 0, 0, 5, 6, 7},
}

type (
	//Card object
	/*
	 * This is the Card Object
	 *   @number: the number representing the card number
	 *   @suit: this is the card suit
	 */
	Card struct {
		number int32
		suit   string
		weight int
	}
)

//NewCard create a card object
func NewCard(num int32, suit string) *Card {
	return &Card{
		number: num,
		suit:   suit,
		weight: weights[suit][num],
	}
}

/*
 *  Print a card
 */
func (c *Card) show() string {
	return strconv.Itoa(int(c.number)) + "-" + c.suit
}

/*
 *  get a card name
 */
func (c *Card) getCardName() string {
	return c.suit + "_" + strconv.Itoa(int(c.number))
}

/*
 * Compares two cards
 *   @card: the card to compare this
 *
 * Returns:
 *   1 if this card is better than 'card',
 *   0 if are equal and
 *   -1 if it's worst
 */
func (c *Card) confront(cc *Card) int {
	if c.weight > cc.weight {
		return 1
	} else if c.weight == cc.weight {
		return 0
	} else if c.weight < cc.weight {
		return -1
	}
	return -2
}

/*
 * Returns the envido points of two cards 'this' and 'card'
 */
func (c *Card) envido(cc *Card) int32 {
	cardValue := cc.number
	if cc.isBlackCard() {
		cardValue = 0
	}
	thisValue := c.number
	if c.isBlackCard() {
		thisValue = 0
	}

	if !c.isSameSuit(cc) {
		return nano.Max(int32(cardValue), int32(thisValue))
	} else if cc.isBlackCard() && c.isBlackCard() {
		return 20
	} else {
		return int32(cardValue + thisValue + 20)
	}
}

func (c *Card) isBlackCard() bool {
	return c.number == 10 || c.number == 11 || c.number == 12
}

func (c *Card) isSameSuit(cc *Card) bool {
	return c.suit == cc.suit
}

