package fightlandlord

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Card struct {
	Card_cseq  int
	Card_value int
	Card_type  string
	Isalloc    bool
}

type Player struct {
	Player_addr string
	//Player_cards []int
	Player_cards   []Card
	Player_name    string
	Player_type    string
	Player_compete string
	Player_power   bool
	Player_id      int
}

type Discards struct {
	PlayerAcard  []Card
	PlayerBcard  []Card
	PlayerCcard  []Card
	Landlordcard []Card
	Luckey       int
}

type Game struct {
	Game_player    []Player
	Game_cards     Discards
	Game_id        int
	Game_lastcards Comcard
	Game_currcards Comcard
	Game_power     int
}

type Comcard struct {
	value     int
	cardstype string
	exist     bool
	length    int
	addr      string
}

var Games []Game

//准备
func Pregame(conn *net.UDPConn, name string, addr string) {
	if len(Games) == 0 {
		var game Game
		var player Player
		player.Player_addr = addr
		player.Player_name = name
		player.Player_type = "farmer"
		player.Player_id = len(game.Game_player)
		game.Game_player = append(game.Game_player, player)
		game.Game_id = len(Games)
		Games = append(Games, game)
	} else {
		k := 0
		for i := 0; i < len(Games); i++ {
			j := len(Games[i].Game_player)
			if j < 3 {
				var player Player
				player.Player_addr = addr
				player.Player_name = name
				player.Player_type = "farmer"
				player.Player_id = len(Games[i].Game_player)
				Games[i].Game_player = append(Games[i].Game_player, player)
				if len(Games[i].Game_player) == 3 {
					Games[i].start(conn)
				}
				break
			}
			k = i
		}
		if k == len(Games) {
			var game Game
			var player Player
			player.Player_addr = addr
			player.Player_name = name
			player.Player_type = "farmer"
			player.Player_id = len(game.Game_player)
			game.Game_player = append(game.Game_player, player)
			game.Game_id = len(Games)
			Games = append(Games, game)
		}
	}
}

//开始
func (game *Game) start(conn *net.UDPConn) {

	for _, player := range game.Game_player {
		fmt.Println(player)
		pAddr, err := net.ResolveUDPAddr("udp", player.Player_addr)
		checkError(err)
		str := "connected" + "\r\n" + "connected three players,let's start..."
		conn.WriteToUDP([]byte(str), pAddr)
	}

	game.Game_cards = distributecard_test()

	game.Game_player[0].Player_cards = game.Game_cards.PlayerAcard
	game.Game_player[1].Player_cards = game.Game_cards.PlayerBcard
	game.Game_player[2].Player_cards = game.Game_cards.PlayerCcard

	game.initsendcard(conn)
	str := "competing"
	cmd := "allocrole"
	game.send(conn, str, cmd)
}

func distributecard_test() Discards {

	cards := initcards()
	//玩家3堆牌，底牌
	var discards Discards

	//抽地主牌：直接选择一名玩家
	rand.Seed(time.Now().Unix())
	discards.Luckey = rand.Intn(3)

	//抽3张底牌
	for i := 0; i < 3; i++ {
		myrand := rand.Intn(54)
		disnum := distribute(myrand, cards)
		var card Card
		card = cards[disnum]
		discards.Landlordcard = append(discards.Landlordcard, card)
	}

	//将剩余51张分为3份
	for i := 0; i < 51; i++ {
		myrand := rand.Intn(54)
		if i%3 == 0 {

			disnum := distribute(myrand, cards)
			var card Card
			card = cards[disnum]
			discards.PlayerAcard = append(discards.PlayerAcard, card)
		} else if i%3 == 1 {

			disnum := distribute(myrand, cards)
			var card Card
			card = cards[disnum]
			discards.PlayerBcard = append(discards.PlayerBcard, card)
		} else if i%3 == 2 {

			disnum := distribute(myrand, cards)
			var card Card
			card = cards[disnum]
			discards.PlayerCcard = append(discards.PlayerCcard, card)
		}
	}
	return discards
}

func (game *Game) initsendcard(conn *net.UDPConn) {

	for i := 0; i < 3; i++ {
		var nums []int
		for _, card := range game.Game_player[i].Player_cards {
			if card.Isalloc == true {
				nums = append(nums, card.Card_value)
			}
		}

		Addr, err := net.ResolveUDPAddr("udp", game.Game_player[i].Player_addr)
		checkError(err)
		str := Intstostring(nums)
		str = "distributecards" + "\r\n" + str
		fmt.Println(str)
		conn.WriteToUDP([]byte(str), Addr)
	}

}

func initcards() []Card {
	var cards [54]Card
	num1 := 1
	num3 := 1
	num2 := 1
	num4 := 1
	for i := 0; i < 54; i++ {
		if i < 13 {
			cards[i].Card_cseq = i
			cards[i].Card_value = num1
			cards[i].Card_type = "black"
			num1++
		} else if i >= 13 && i < 26 {
			cards[i].Card_cseq = i
			cards[i].Card_value = num2
			cards[i].Card_type = "red"
			num2++
		} else if i >= 26 && i < 39 {
			cards[i].Card_cseq = i
			cards[i].Card_value = num3
			cards[i].Card_type = "plum"
			num3++
		} else if i >= 39 && i < 52 {
			cards[i].Card_cseq = i
			cards[i].Card_value = num4
			cards[i].Card_type = "diamond"
			num4++
		} else if i == 52 {
			cards[i].Card_cseq = i
			cards[i].Card_value = 14
			cards[i].Card_type = "small"
		} else if i == 53 {
			cards[i].Card_cseq = i
			cards[i].Card_value = 15
			cards[i].Card_type = "big"
		}
	}
	var outcards []Card = cards[:]
	return outcards
}

func main() {

	service := ":8050"
	udpAddr, err := net.ResolveUDPAddr("udp", service)
	checkError(err)
	conn, err := net.ListenUDP("udp", udpAddr)
	checkError(err)
	fmt.Println("Server start ...")
	for {
		handleClientudp(conn)
	}

}

func handleClientudp(conn *net.UDPConn) {
	var buf [100]byte
	//var player Player
	n, addr, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		return
	}

	buffer := string(buf[:n])
	packetinfo := strings.Split(buffer, "\r\n")
	if len(packetinfo) != 2 {
		fmt.Println("the packet not belong us...")
	}
	if packetinfo[0] == "connected" {
		fmt.Println("connected sucess...")
		Pregame(conn, packetinfo[1], addr.String())
		//fmt.Println(Games)
	} else if packetinfo[0] == "compete" {
		//fmt.Println("获得抢地主")
		competelandlord(conn, packetinfo[1], addr.String())

	} else if packetinfo[0] == "compare" {
		fmt.Println(packetinfo[1])
		comparecards_test(conn, packetinfo[1], addr.String())
	} else if packetinfo[0] == "again" {

	} else if packetinfo[0] == "end" {

	}
}

func getbelonggameid(addr string) (i int, j int) {

	for i = 0; i < len(Games); i++ {
		for j = 0; j < len(Games[i].Game_player); j++ {
			if Games[i].Game_player[j].Player_addr == addr {
				return i, j
			}
		}
	}
	return -1, -1
}

func comparecards_test(conn *net.UDPConn, str string, addr string) {
	gameid, playerid := getbelonggameid(addr)
	nums := stringtoints(str)
	sort.Ints(nums)
	fmt.Println("zhelishicompare")
	fmt.Println(nums)

	//是否出牌
	if len(nums) == 1 && nums[0] == 0 {
		nextid := (playerid + 1) % 3
		if Games[gameid].Game_lastcards.addr == Games[gameid].Game_player[nextid].Player_addr {
			Games[gameid].Game_power++
			getcards(conn, gameid, "allocpower")
		} else {
			Games[gameid].Game_power++
			getcards(conn, gameid, "select")
		}

	} else {
		suicards, effective := cardsiseffective_test(gameid, nums, addr)
		//牌是否存在
		if effective {
			repeatnum, value := getrepeatnum(nums)
			cardstype := getcardstype(repeatnum, nums)
			//牌类型是否正确
			if cardstype != "err" {

				if len(suicards) == 0 {
					//fmt.Println("本轮游戏结束")
					parecmd := "info"
					parestr := "本轮游戏结束，玩家" + Games[gameid].Game_player[playerid].Player_name + "获得游戏胜利"
					Games[gameid].send(conn, parestr, parecmd)

				} else {
					Games[gameid].Game_currcards.addr = addr
					Games[gameid].Game_currcards.cardstype = cardstype
					Games[gameid].Game_currcards.value = value
					Games[gameid].Game_currcards.exist = true
					Games[gameid].Game_currcards.length = len(nums)

					parecmd := "info"
					parestr := "玩家" + Games[gameid].Game_player[playerid].Player_name + "出牌：" + str
					Games[gameid].send(conn, parestr, parecmd)

					//上牌是否存在
					if Games[gameid].Game_lastcards.exist {
						if Games[gameid].Game_lastcards.addr == addr {
							Games[gameid].Game_lastcards = Games[gameid].Game_currcards
							Games[gameid].Game_player[playerid].Player_cards = suicards
							//下家选择出牌
							Games[gameid].Game_power++
							getcards(conn, gameid, "select")
						} else {
							if Games[gameid].compare() {
								Games[gameid].Game_lastcards = Games[gameid].Game_currcards
								Games[gameid].Game_player[playerid].Player_cards = suicards
								//下家选择出牌
								Games[gameid].Game_power++
								getcards(conn, gameid, "select")
							} else {
								//重新出牌
								getcards(conn, gameid, "select")
							}
						}
					} else {
						Games[gameid].Game_lastcards = Games[gameid].Game_currcards
						Games[gameid].Game_player[playerid].Player_cards = suicards
						//下家选择出牌
						Games[gameid].Game_power++
						getcards(conn, gameid, "select")
					}
				}

			} else {
				//重新出牌
				if Games[gameid].Game_player[playerid].Player_power {
					getcards(conn, gameid, "allocpower")
				} else {
					getcards(conn, gameid, "select")
				}
			}
		} else {
			//重新出牌
			if Games[gameid].Game_player[playerid].Player_power {
				getcards(conn, gameid, "allocpower")
			} else {
				getcards(conn, gameid, "select")
			}
		}
	}
}

func getcards(conn *net.UDPConn, i int, cmdstr string) {

	Games[i].initsendcard(conn)

	cseq := Games[i].Game_power % 3
	//Games[i].Game_power = cseq + 1
	cmd := "info"
	str := "玩家" + Games[i].Game_player[cseq].Player_name + "拥有牌权，请出牌："
	Games[i].send(conn, str, cmd)

	Games[i].sendpower(conn, cmdstr, Games[i].Game_player[cseq].Player_addr)
}

func (game *Game) compare() bool {
	if game.Game_currcards.cardstype == "bomb" {
		if game.Game_lastcards.cardstype != "bomb" {
			return true
		} else if game.Game_lastcards.value < game.Game_currcards.value {
			return true
		} else {
			return false
		}
	} else {
		if game.Game_currcards.cardstype == game.Game_lastcards.cardstype && game.Game_currcards.length == game.Game_lastcards.length {
			if game.Game_currcards.value > game.Game_lastcards.value {
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	}
	return false
}

func cardsiseffective_test(gameid int, nums []int, addr string) ([]Card, bool) {
	var cards []Card
	var surcards []Card
	for i := 0; i < len(Games[gameid].Game_player); i++ {
		if Games[gameid].Game_player[i].Player_addr == addr {
			for _, card := range Games[gameid].Game_player[i].Player_cards {
				cards = append(cards, card)
			}
		}
	}
	if len(nums) > len(cards) {
		return cards, false
	}
	k := 0
	for m := 0; m < len(nums); m++ {
		for n := 0; n < len(cards); n++ {
			if cards[n].Card_value == nums[m] && cards[n].Isalloc == true {
				cards[n].Isalloc = false
				k = n
				break
			}
		}
		if k == len(cards) {
			return cards, false
		}
	}
	for _, card := range cards {
		if card.Isalloc == true {
			surcards = append(surcards, card)
		}
	}

	return surcards, true
}

//获取重复数个数及其值
func getrepeatnum(nums []int) (int, int) {
	max := 1
	repeat := 1
	value := nums[0]
	if len(nums) == 1 {
		return 1, nums[0]
	} else if len(nums) == 2 && nums[0] == 14 && nums[1] == 15 || len(nums) == 2 && nums[1] == 14 && nums[0] == 15 {
		return 4, nums[0]
	}
	for i := 1; i < len(nums); i++ {
		if nums[i] == nums[i-1] {
			repeat = repeat + 1
			if repeat > max {
				max = repeat
				value = nums[i]
			}
		} else {
			repeat = 1
		}
	}
	fmt.Println("重复次数", max, "重复值", value)
	return max, value
}

//获取出牌类型
func getcardstype(repeatnum int, nums []int) string {
	if repeatnum == 1 {
		if len(nums) == 1 {
			str := "single"
			return str
		} else if len(nums) >= 5 && iscsequence(nums) {
			str := "sequence"
			return str
		} else {
			str := "err"
			return str
		}
	} else if repeatnum == 2 {
		if len(nums) == 2 {
			str := "double"
			return str
		} else if len(nums) >= 6 && len(nums)%2 == 0 && iscsequencepair(nums) {
			str := "sequencepair"
			return str
		} else {
			str := "err"
			return str
		}
	} else if repeatnum == 3 {
		if len(nums) == 3 {
			str := "three"
			return str
		} else if len(nums) == 4 {
			str := "threeandone"
			return str
		} else {
			str := "err"
			return str
		}
	} else if repeatnum == 4 {
		if len(nums) == 4 || len(nums) == 2 {
			str := "bomb"
			return str
		} else if len(nums) == 6 {
			str := "fourandtwo"
			return str
		} else {
			str := "err"
			return str
		}
	}
	str := "err"
	return str
}

//连牌
func iscsequence(nums []int) bool {
	for i := 1; i < len(nums); i++ {
		if nums[i]-nums[i-1] != 1 {
			return false
		}
	}
	return true
}

//连对
func iscsequencepair(nums []int) bool {
	for i := 2; i < len(nums); i++ {
		if nums[i]-nums[i-2] != 1 {
			return false
		}
		i = i + 1
	}
	return true
}

func stringtoints(str string) []int {
	var nums []int
	strs := strings.Split(str, ",")
	for i := 0; i < len(strs); i++ {
		var num int
		num, _ = strconv.Atoi(strs[i])
		nums = append(nums, num)
	}
	fmt.Println(nums)
	return nums
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}

func distribute(disnum int, cards []Card) int {
	//fmt.Println(cards)
	if disnum > 53 {
		disnum = disnum - 53
	}
	if cards[disnum].Isalloc == false {
		cards[disnum].Isalloc = true
		return disnum
	} else {
		return distribute(disnum+1, cards)
	}
}

func Intstostring(nums []int) string {
	var str string
	for _, num := range nums {
		str = str + strconv.Itoa(num) + ","
	}
	//fmt.Println(str)
	return str
}

func setcompete(str string, addr string) int {

	for i := 0; i < len(Games); i++ {
		for j := 0; j < len(Games[i].Game_player); j++ {
			if Games[i].Game_player[j].Player_addr == addr {
				Games[i].Game_player[j].Player_compete = str
				return i
			}
		}
	}
	return -1

}

func competelandlord(conn *net.UDPConn, str string, addr string) {
	num := setcompete(str, addr)
	//fmt.Println(Games[num])
	if Games[num].Game_player[0].Player_compete != "" && Games[num].Game_player[1].Player_compete != "" && Games[num].Game_player[2].Player_compete != "" {
		a, _ := strconv.Atoi(Games[num].Game_player[0].Player_compete)
		b, _ := strconv.Atoi(Games[num].Game_player[1].Player_compete)
		c, _ := strconv.Atoi(Games[num].Game_player[2].Player_compete)
		if Games[num].Game_cards.Luckey == 0 {
			a = a + 2
			b = b + 1
		} else if Games[num].Game_cards.Luckey == 1 {
			b = b + 2
			c = c + 1
		} else if Games[num].Game_cards.Luckey == 2 {
			c = c + 2
			a = a + 1
		}
		cseq := getmax(a, b, c)
		Landlord := "landlord"
		Games[num].Game_player[cseq].Player_type = Landlord
		Games[num].Game_player[cseq].Player_power = true
		Games[num].Game_power = cseq
		for _, card := range Games[num].Game_cards.Landlordcard {
			Games[num].Game_player[cseq].Player_cards = append(Games[num].Game_player[cseq].Player_cards, card)
		}
		//fmt.Println(Games[num].Game_player[cseq].Player_cards)
		str := "竞争地主完毕,玩家" + Games[num].Game_player[cseq].Player_name + "成为地主."
		cmd := "info"
		Games[num].send(conn, str, cmd)
		Games[num].initsendcard(conn)

		cmd = "info"
		str = "玩家" + Games[num].Game_player[cseq].Player_name + "拥有牌权，请出牌："
		Games[num].send(conn, str, cmd)

		cmd = "allocpower"
		Games[num].sendpower(conn, cmd, Games[num].Game_player[cseq].Player_addr)
	}
}

func getmax(a int, b int, c int) int {
	if a > b && a > c {
		return 0
	} else if b > a && b > c {
		return 1
	} else if c > a && c > b {
		return 2
	}
	return -1
}

func (game *Game) sendpower(conn *net.UDPConn, cmd string, addr string) {
	Addr, err := net.ResolveUDPAddr("udp", addr)
	checkError(err)
	sendstr := cmd + "\r\n"
	conn.WriteToUDP([]byte(sendstr), Addr)
}

func (game *Game) send(conn *net.UDPConn, str string, cmd string) {

	for i := 0; i < 3; i++ {

		Addr, err := net.ResolveUDPAddr("udp", game.Game_player[i].Player_addr)
		checkError(err)
		sendstr := cmd + "\r\n" + str
		//fmt.Println(sendstr)
		conn.WriteToUDP([]byte(sendstr), Addr)

	}
}
