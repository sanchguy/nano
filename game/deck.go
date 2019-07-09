package game

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

/*
* Suit of Spanish deck
 */
var suits = []string{"oro", "copa", "espada", "basto"}

/*
* Number of cards used in Truco game
 */
var carNumbers = []int{1, 2, 3, 4, 5, 6, 7, 11, 12}

type (
	//Deck constructor
	Deck struct {
	}
)

//NewDeck return a deck object
func NewDeck() *Deck {
	return &Deck{}
}

func (d *Deck) sorted() []*Card {
	var deck []*Card
	for _, suit := range suits {
		for _, cardNumber := range carNumbers {
			deck = append(deck, NewCard(cardNumber, suit))
		}
	}
	fmt.Print(deck)
	rand.Seed(time.Now().Unix())
	d.random(deck, 18)
	fmt.Print("random cards", deck)
	return deck
}

//Random deck cards
func (d *Deck) random(card []*Card, length int) {
	if len(card) <= 0 {
		fmt.Print(errors.New("the length of the parameter strings should not be less than 0"))
	}

	if length <= 0 || len(card) <= length {
		fmt.Print(errors.New("the size of the parameter length illegal"))
	}

	for i := len(card) - 1; i > 0; i-- {
		num := rand.Intn(i + 1)
		card[i], card[num] = card[num], card[i]
	}
}
