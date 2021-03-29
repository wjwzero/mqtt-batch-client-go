package mqtt_cluster_ip

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

/**
从南向获得分配到的mqtt节点
*/
func GetMqttClusterIp(southUrl string, deviceKey string) string {
BEGIN:
	resp, err := http.Get(southUrl + "&deviceKey=" + deviceKey)
	if err != nil {
		log.Print("访问中心节点失败", err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	str2 := string(body)
	fmt.Printf("body http %s", str2)
	json, err := simplejson.NewJson(body)
	code, err := json.Get("code").Int()
	log.Print(code)
	var res string
	if code == 200 {
		res, _ = json.Get("data").Get("ip").String()
	} else {
		res, _ = json.Get("data").String()
		log.Printf("获得cluster异常%s", res)
		time.Sleep(1 * time.Second)
		goto BEGIN
	}
	return res
}
