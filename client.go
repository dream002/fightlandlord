package client

import (
	//"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	player := flag.String("n", "p3", "玩家id")
	server := flag.String("dst", "172.25.202.180:8050", "服务器地址")
	service := flag.String("src", ":8053", "本地地址")
	flag.Parse()
	//地址解析
	udpAddr, err := net.ResolveUDPAddr("udp", *service)
	checkError(err)
	maddr, err := net.ResolveUDPAddr("udp", *server)
	checkError(err)
	//监听端口
	conn, err := net.ListenUDP("udp", udpAddr)
	checkError(err)
	//将玩家id发给服务器
	str := *player
	str = "connected" + "\r\n" + str
	_, err = conn.WriteToUDP([]byte(str), maddr)
	checkError(err)
	//获取socket消息
	for {
		handleClientudp(conn)
	}
}

func handleClientudp(conn *net.UDPConn) {

	var buf [100]byte
	n, _, err := conn.ReadFromUDP(buf[0:])
	if err != nil {
		return
	}
	//fmt.Println(string(buf[:n]))

	buffer := string(buf[:n])
	packetinfo := strings.Split(buffer, "\r\n")
	if len(packetinfo) != 2 {
		fmt.Println("the packet not belong us...")
	}
	if packetinfo[0] == "connected" {
		fmt.Println("connected server...")
	} else if packetinfo[0] == "info" {
		fmt.Println(packetinfo[1])
	} else if packetinfo[0] == "distributecards" {
		getcards(packetinfo[1], conn)
	} else if packetinfo[0] == "allocrole" {
		compete(conn)
	} else if packetinfo[0] == "allocpower" {
		sendmycards(conn)
	} else if packetinfo[0] == "select" {
		selectsendcards(conn)
	} else if packetinfo[0] == "judge" {

	} else if packetinfo[0] == "end" {

	}

}

func selectsendcards(conn *net.UDPConn) {
	var str string
	fmt.Println("请出牌，牌间用逗号隔开，输入0表示pass")
	fmt.Scanln(&str)
	cmd := "compare"
	Sendmsg(cmd, str, conn)
}

func sendmycards(conn *net.UDPConn) {
	var str string
	fmt.Println("请出牌，牌间用逗号隔开")
	fmt.Scanln(&str)
	k := []byte(str)
	for len(k) == 1 && string(k[0]) == "0" {
		fmt.Println("必须出牌，牌间用逗号隔开")
		fmt.Scanln(&str)
		k = []byte(str)
	}
	cmd := "compare"
	Sendmsg(cmd, str, conn)
}

func getcards(card string, conn *net.UDPConn) {

	cards := stringtoints(card)
	sort.Ints(cards)
	fmt.Println("get the cards:")
	fmt.Println(cards)

	//Sendmsg(str, conn)
}

func compete(conn *net.UDPConn) {
	var cmdcompetestr string
	fmt.Println("是否抢地主（输入0不抢，输入3抢）：")
	fmt.Scanln(&cmdcompetestr)
	/*for cmdcompete != 0 && cmdcompete != 3 {
		fmt.Println("输入错误，请重新输入。")
		fmt.Println("是否抢地主（输入0不抢，输入3抢）：")
		fmt.Scanf("%d", &cmdcompete)
	}*/
	cmdcompete, _ := strconv.Atoi(cmdcompetestr)
	if cmdcompete == 0 || cmdcompete == 3 {
		//cmdcompetestr := strconv.Itoa(cmdcompete)
		str := "compete" + "\r\n" + cmdcompetestr
		server := "172.25.202.180:8050"
		maddr, err := net.ResolveUDPAddr("udp", server)
		checkError(err)
		_, err = conn.WriteToUDP([]byte(str), maddr)
		checkError(err)
	}
}

func Sendmsg(cmd string, msg string, conn *net.UDPConn) {
	str := cmd + "\r\n" + msg
	server := "172.25.202.180:8050"
	maddr, err := net.ResolveUDPAddr("udp", server)
	checkError(err)
	_, err = conn.WriteToUDP([]byte(str), maddr)
	checkError(err)
}

func stringtoints(str string) []int {
	var nums []int
	strs := strings.Split(str, ",")
	for i := 0; i < len(strs)-1; i++ {
		var num int
		num, _ = strconv.Atoi(strs[i])
		nums = append(nums, num)
	}
	//fmt.Println(nums)
	return nums
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}
