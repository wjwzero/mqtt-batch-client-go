package main

import (
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var downCron = flag.String("downCron", "*/1 * * * * *", "cron表达式, 默认每秒")
var downNum = flag.Int("downNum", 3000, "向多少个终端发送消息")
var url = flag.String("url", "http://admin.skyworthtest.top/api", "地址")
var random = flag.Int("random", 3, "cron后随机延时发送时间,单位s")
var payload = flag.String("payload", "testtest", "cron后随机延时发送时间,单位毫秒")

func init() {
	flag.Parse()
	getToken()
	log.Printf("downNum %d", *downNum)
}

func command() {
	dks := ""
	for i := 1; i <= *downNum; i++ {
		/*log.Printf("Num %d", i)
		if *random != 0 {
			num := rand.Intn(*random)
			log.Printf("延时时间 %d", num)
			time.Sleep(time.Duration(num) * time.Millisecond)
		}*/
		if i != 1 && i != 1001 && i != 2001 {
			dks += ","
		}
		dks += "\"1test" + strconv.Itoa(i) + "\""

		if i%1000 == 0 {
			log.Printf("dks : %s", dks)
			HttpIotCommand(dks)
			dks = ""
		}
	}
}

func main() {
	/*cr := cron.New()
	if *downCron == "X" {
		command()
		select {}
	}
	_, _ = cr.AddFunc(*downCron, command)
	cr.Start()
	select {}*/
	getToken()
}

var token string

func HttpIotCommand(dk string) {
	url := *url + "/skyworthNorthbound/v1/MQTTSendMessage/sendAliasListMessage"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
		"aliasList": [`+dk+`],
		"aliasType": "deviceId",
		"appKey": "57634393",
		"appName": "test",
		"appSecret": "10508e45e951ac22",
		"payload": "%v",
		"productKey": "bench_mqtt",
		"retain": false,
		"messageQueryId": "1",
		"messageType": "test",
		"test": false
	}`, *payload))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	log.Print(token)
	req.Header.Add("Blade-Auth", "bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}

func getToken() {

	url := *url + "/skyworthNorthbound/user/devLogin"
	method := "POST"

	payload := strings.NewReader(`{
    "tenantId": "000000",
    "username": "18729891158",
    "password": "admin",
    "grant_type": "password",
    "scope": "all",
    "type": "account"
	}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "Basic YmxhZGV1c2VyOmJsYWRldXNlcl9zZWNyZXQ=")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	json, _ := simplejson.NewJson(body)
	token, _ = json.Get("data").Get("access_token").String()
}
