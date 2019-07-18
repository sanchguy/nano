package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/looplab/fsm"
)

var posiblesE = []map[string]string{
	{"p": "envido"}, {"p": "envido-envido"}, {"p": "real-envido"}, {"p": "envido-real"},
	{"p": "envido-envido-real"}, {"p": "falta-envido"}, {"p": "envido-falta"}, {"p": "envido-envido-falta"},
	{"p": "envido-envido-real-falta"}, {"p": "real-envido-falta"}, {"p": "envido-real-falta"},
}

var tablep1 = []*Card{}
var tablep2 = []*Card{}

type (
	//Round1 is one round in games object
	Round1 struct {
		player1name int64

		player2name int64

		currentHand int64

		currentTrun int64

		FSM *fsm.FSM

		status string

		score []int32

		turnWin []int

		tablep1 []*Card

		tablep2 []*Card

		flagTruco bool

		flagRetruco bool

		flagValeCuatro bool

		flagNoCanto bool

		auxWin bool

		cartasp1 []*Card

		cartasp2 []*Card

		pointsEnvidoP1 int

		pointsEnvidoP2 int

		pardas bool
	}
)

//NewRound return a round obj
func NewRound(room *Room) *Round1 {
	r := &Round1{
		player1name: room.players[0].id,

		player2name: room.players[1].id,

		currentHand: room.currentHand,

		currentTrun: room.currentTurn,

		status: "running",

		score: []int32{},

		turnWin: []int{},

		tablep1: []*Card{},

		tablep2: []*Card{},

		flagTruco: false,

		flagRetruco: false,

		flagValeCuatro: false,

		flagNoCanto: false,

		auxWin: false,

		cartasp1: []*Card{},

		cartasp2: []*Card{},

		pointsEnvidoP1: 0,

		pointsEnvidoP2: 0,

		pardas: false,
	}

	r.FSM = r.newTrucoFSM(room.currentState)
	return r
}

func (r *Round1) newTrucoFSM(state string) *fsm.FSM {
	initialState := ""
	if state == "" {
		initialState = "init"
	}
	trucoFsm := fsm.NewFSM(initialState, fsm.Events{
		{Name: "playcard", Src: []string{"init"}, Dst: "primer-carta"},
		{Name: "envido", Src: []string{"init", "primer-carta"}, Dst: "envido"},
		{Name: "envido-envido", Src: []string{"envido"}, Dst: "envido-envido"},
		{Name: "envido-real", Src: []string{"envido"}, Dst: "envido-real"},
		{Name: "envido-envido-real", Src: []string{"envido-envido"}, Dst: "envido-envido-real"},
		{Name: "real-envido", Src: []string{"init", "primer-carta"}, Dst: "real-envido"},
		{Name: "falta-envido", Src: []string{"init", "primer-carta", "envido", "envido-envido",
			"real-envido", "envido-envido-real", "envido-real"}, Dst: "falta-envido"},
		{Name: "envido-falta", Src: []string{"envido"}, Dst: "envido-falta"},
		{Name: "envido-envido-falta", Src: []string{"envido-envido"}, Dst: "envido-envido-falta"},
		{Name: "envido-real-falta", Src: []string{"envido-real"}, Dst: "envido-real-falta"},
		{Name: "envido-envido-real-falta", Src: []string{"envido-envido-real"}, Dst: "envido-envido-real-falta"},
		{Name: "real-envido-falta", Src: []string{"real-envido"}, Dst: "real-envido-falta"},
		{Name: "truco", Src: []string{"init", "played-card",
			"playcard", "primer-carta",
			"quiero", "no-quiero"}, Dst: "truco"},
		{Name: "retruco", Src: []string{"truco", "quiero", "playcard", "played-card"}, Dst: "retruco"},
		{Name: "valecuatro", Src: []string{"retruco", "quiero", "playcard", "played-card"}, Dst: "valecuatro"},
		{Name: "playcard", Src: []string{"quiero", "no-quiero",
			"primer-carta", "played-card",
			"envido", "truco", "retruco", "valecuatro"}, Dst: "played-card"},
		{Name: "quiero", Src: []string{"envido", "envido-envido", "envido-envido-real", "envido-real",
			"real-envido", "real-envido", "falta-envido",
			"envido-falta", "envido-envido-falta",
			"envido-real-falta", "envido-envido-real-falta",
			"real-envido-falta", "truco", "retruco", "valecuatro"}, Dst: "quiero"},
		{Name: "no-quiero", Src: []string{"envido", "envido-envido", "envido-envido-real", "envido-real",
			"real-envido", "real-envido", "falta-envido",
			"envido-falta", "envido-envido-falta",
			"envido-real-falta", "envido-envido-real-falta",
			"real-envido-falta", "truco", "retruco", "valecuatro"}, Dst: "no-quiero"},
	}, fsm.Callbacks{
		"enter_state": func(e *fsm.Event) { fmt.Printf("fsm enter_state to %s is %s", e.Src, e.Dst) },
	})
	return trucoFsm
}

func (r *Round1) switchPlayer(pname int64) int64 {
	if r.player1name == pname {
		return r.player2name
	}
	return r.player1name

}

func (r *Round1) returnSuit(value string) string {
	return strings.Split(value, "-")[1]
}

func (r *Round1) returnNumber(value string) int {
	num, err := strconv.Atoi(strings.Split(value, "-")[0])
	if err != nil {
		fmt.Println(err)
	}
	return num
}

func (r *Round1) returnValueComplete(value string) string {
	s := ""
	s += strconv.Itoa(r.returnNumber(value))
	s += r.returnSuit(value)
	return s
}

func (r *Round1) actionCurrent() string {
	return r.FSM.Current()
}

func (r *Round1) actionPrevious() string {
	return r.FSM.Current()
}

func (r *Round1) distHamming(arr1 []int, arr2 []int) int {
	if len(arr1) != len(arr2) {
		return -1
	}
	counter := 0
	for index, value := range arr1 {
		if value != arr2[index] {
			counter++
		}
	}
	return counter
}

func (r *Round1) confrontScore() {
	switch len(r.turnWin) {
	case 0:
		if tablep1[0] != nil && tablep2[0] != nil {
			card1 := r.tablep1[0]
			card2 := r.tablep2[0]
			conf := card1.confront(card2)
			r.selectWin(conf)
		}
		break
	case 1:
		break
	case 2:
		break
	}
}

func (r *Round1) selectWin(conf int) []int {
	switch conf {
	case -1:
		r.turnWin = append(r.turnWin, 1)
		return r.turnWin
	case 0:
		r.turnWin = append(r.turnWin, -1)
		return r.turnWin
	case 1:
		r.turnWin = append(r.turnWin, 0)
		return r.turnWin
	}
	return r.turnWin
}

func (r *Round1) changeTurn() int64 {
	if len(r.tablep1) != len(r.tablep2) || r.FSM.Current() == "truco" || r.FSM.Current() == "retruco" || r.FSM.Current() == "valecuatro" {
		r.currentTrun = r.switchPlayer(r.currentTrun)
		return r.currentTrun
	}
	if len(r.turnWin) != 0 {
		switch r.turnWin[len(r.turnWin)-1] {
		case 0:
			r.currentTrun = r.player1name
			return r.currentTrun
		case 1:
			r.currentTrun = r.player2name
			return r.currentTrun
		case -1:
			r.currentTrun = r.currentHand
			return r.currentTrun
		}
	}
	r.currentTrun = r.switchPlayer(r.currentTrun)
	return r.currentTrun
}

func (r *Round1) findPosiblesE(p string, action string) bool {
	foundActon := false
	for _, actionMap := range posiblesE {
		for key, value := range actionMap {
			if key == p {
				if value == action {
					foundActon = true
					return foundActon
				}
			}
		}
	}
	return foundActon
}

func (r *Round1) calculateScore(action string, prev string, value string, playerid int64) {
	if (action == "played-card" || action == "playcard") && r.auxWin == false {
		r.setTable(value, playerid)
		r.confrontScore()
		if (r.flagTruco == true) && (len(r.tablep1) > 1) && (len(r.tablep2) > 1) {
			r.calculateScoreTruco(action, playerid, "")
		}
		if (r.flagTruco == false) && (r.flagNoCanto == false) && (len(r.tablep1) > 1 && len(r.tablep2) > 1) {
			r.calculateScoreTruco(action, playerid, "")
		}
	}
	if (action == "quiero" || action == "no-quiero") && (r.findPosiblesE("p", prev) == true) {
		r.calculateScoreEnvido(action, prev, playerid)
	}
	if (action == "quiero" || action == "no-quiero") && prev == "truco" {
		if action == "quiero" {
			r.flagTruco = true
		}
		if (len(r.tablep1) < 2) || (len(r.tablep2) < 2) {
			r.calculateScoreTruco(action, playerid, "")
		}
	}
	if (action == "quiero" || action == "no-quiero") && prev == "retruco" {
		r.flagRetruco = true
		if len(r.tablep1) < 2 || len(r.tablep2) < 2 {
			r.calculateScoreTruco(action, playerid, value)
		}
	}
	if (action == "quiero" || action == "no-quiero") && prev == "valecuatro" {
		r.flagValeCuatro = true
		if len(r.tablep1) < 2 || len(r.tablep2) < 2 {
			r.calculateScoreTruco(action, playerid, value)
		}
	}
}

func (r *Round1) calculateScoreEnvido(action string, prev string, player int64) {
	var total int32 = 9
	if action == "quiero" {
		switch prev {
		case "envido":
			r.assignPoints(action, 2, player)
			break
		case "real-envido":
			r.assignPoints(action, 3, player)
			break
		case "envido-envido":
			r.assignPoints(action, 4, player)
			break
		case "envido-real":
			r.assignPoints(action, 5, player)
			break
		case "envido-envido-real":
			r.assignPoints(action, 7, player)
			break
		case "falta-envido":
			if player == r.player1name {
				r.assignPoints(action, total-(r.score[1]), player)
			}

			if player == r.player2name {
				r.assignPoints(action, total-(r.score[0]), player)
			}
			break
		case "envido-falta":
			if player == r.player1name {
				r.assignPoints(action, total-(r.score[1]), player)
			}

			if player == r.player2name {
				r.assignPoints(action, total-(r.score[0]), player)
			}
			break
		case "envido-real-falta":
			if player == r.player1name {
				r.assignPoints(action, total-(r.score[1]), player)
			}

			if player == r.player2name {
				r.assignPoints(action, total-(r.score[0]), player)
			}
			break
		case "envido-envido-falta":
			if player == r.player1name {
				r.assignPoints(action, total-(r.score[1]), player)
			}

			if player == r.player2name {
				r.assignPoints(action, total-(r.score[0]), player)
			}
			break
		case "envido-envido-real-falta":
			if player == r.player1name {
				r.assignPoints(action, total-(r.score[1]), player)
			}

			if player == r.player2name {
				r.assignPoints(action, total-(r.score[0]), player)
			}
			break
		case "real-envido-falta":
			if player == r.player1name {
				r.assignPoints(action, total-(r.score[1]), player)
			}

			if player == r.player2name {
				r.assignPoints(action, total-(r.score[0]), player)
			}
			break
		}
	}

	if action == "no-quiero" {
		switch prev {
		case "envido":
			r.assignPoints(action, 1, player)
			break
		case "real-envido":
			r.assignPoints(action, 1, player)
			break
		case "envido-envido":
			r.assignPoints(action, 2, player)
			break
		case "envido-real":
			r.assignPoints(action, 2, player)
			break
		case "envido-envido-real":
			r.assignPoints(action, 4, player)
			break
		case "falta-envido":
			r.assignPoints(action, 1, player)
			break
		case "envido-falta":
			r.assignPoints(action, 2, player)
			break
		case "envido-real-falta":
			r.assignPoints(action, 5, player)
			break
		case "envido-envido-falta":
			r.assignPoints(action, 4, player)
			break
		case "envido-envido-real-falta":
			r.assignPoints(action, 7, player)
			break
		case "real-envido-falta":
			r.assignPoints(action, 3, player)
			break
		}
	}
}

func (r *Round1) assignPoints(action string, num int32, playername int64) {
	if action == "quiero" {
		if r.pointsEnvidoP1 > r.pointsEnvidoP2 {
			if r.currentHand == r.player1name {
				r.score[0] += num
			} else {
				r.score[1] += num
			}
		}
		if r.pointsEnvidoP2 > r.pointsEnvidoP1 {
			if r.currentHand == r.player2name {
				r.score[0] += num
			} else {
				r.score[1] += num
			}
		}
		if r.pointsEnvidoP1 == r.pointsEnvidoP2 && r.currentHand == r.player1name {
			r.score[0] += num
		}
		if r.pointsEnvidoP1 == r.pointsEnvidoP2 && r.currentHand == r.player2name {
			r.score[1] += num
		}
	}
	if action == "no-quiero" {
		if playername == r.player1name {
			r.score[1] += num
		}
		if playername == r.player2name {
			r.score[0] += num
		}
	}
}

func (r *Round1) checkWinner(arr [][]int, num int) {
	i := 0
	for {
		if i > len(arr) {
			break
		} else {
			if i < len(arr) && r.auxWin == false {
				elem := arr[i]
				if r.distHamming(elem, r.turnWin) == 0 {
					r.auxWin = true
					if r.flagValeCuatro == true {
						r.score[num] += 4
					} else {
						if r.flagRetruco == true {
							r.score[num] += 3
						} else {
							if r.flagTruco == true {
								r.score[num] += 2
							} else {
								r.score[num]++
								r.flagNoCanto = true
							}
						}
					}
				}
			}
		}
		i++
	}
}

func (r *Round1) calculateScoreTruco(action string, playerid int64, value string) {
	if (action == "quiero" || action == "playcard") && (r.auxWin == false) {
		//posibilodades ganar player1
		var fst = [][]int{{0, 0}, {-1, 0}, {1, 0, 0}, {0, -1}, {-1, -1, 0}, {0, 1, 0}, {0, 1, -1}}
		//posibilodades ganar player2
		var snd = [][]int{{1, 1}, {-1, 1}, {0, 1, 1}, {1, -1}, {-1, -1, 1}, {1, 0, 1}, {1, 0, -1}}
		//posibilidad de triple empate
		var ch = []int{-1, -1, -1}

		if r.distHamming(ch, r.turnWin) == 0 {
			r.calculateScoreTruco(action, playerid, value)
			if r.player1name == r.currentHand {
				r.score[0] += 2
			} else {
				r.score[1] += 2
			}
			r.auxWin = true
		} else {
			r.checkWinner(fst, 0)
			r.checkWinner(snd, 1)
		}
	}

	if action == "no-quiero" {
		r.auxWin = true
		if playerid == r.player1name {
			r.score[1]++
		}
		if playerid == r.player2name {
			r.score[0]++
		}
		if r.flagRetruco == true {
			if playerid == r.player1name {
				r.score[1]++
			}
			if playerid == r.player2name {
				r.score[0]++
			}
		}
		if r.flagValeCuatro == true {
			if playerid == r.player1name {
				r.score[1] += 2
			}
			if playerid == r.player2name {
				r.score[0] += 2
			}
		}
	}
}

func (r *Round1) setTable(value string, player int64) {
	// encontrado := false
	if player == r.player1name {
		card := NewCard(r.returnNumber(value), r.returnSuit(value))
		aux := -1
		i := 0
		for {
			if i > len(r.cartasp1) {
				break
			} else {
				if i < len(r.cartasp1) {
					if r.cartasp1[i].number == card.number && r.cartasp1[i].suit == card.suit {
						aux = i
						// encontrado = true
						r.tablep1 = append(r.tablep1, card)
					}
					i++
				}
			}
		}
		if aux != -1 {
			r.cartasp1 = append(r.cartasp1[:i], r.cartasp1[i+1:]...)
		}
	}

	if player == r.player2name {
		card := NewCard(r.returnNumber(value), r.returnSuit(value))
		aux := -1
		i := 0
		for {
			if i > len(r.cartasp2) {
				break
			} else {
				if i < len(r.cartasp2) {
					if r.cartasp2[i].number == card.number && r.cartasp2[i].suit == card.suit {
						aux = i
						// encontrado = true
						r.tablep2 = append(r.tablep2, card)
					}
					i++
				}
			}
		}
		if aux != -1 {
			r.cartasp2 = append(r.cartasp2[:i], r.cartasp2[i+1:]...)
		}
	}
}

func (r *Round1) play(player int64, action string, value string) *Round1{
	prev := r.actionPrevious()
	r.FSM.Event(action)
	r.calculateScore(action, prev, value, player)
	r.changeTurn()
	return r
}
