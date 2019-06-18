package model

type(
	//Player object
	Player struct{
		id int
		nickname string
		envidoPoints int
		cards []*Card
	}
)

func NewPlayer() *Player{
	
}