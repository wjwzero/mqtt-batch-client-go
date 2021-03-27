package mqtt_cluster_ip

import (
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/hashicorp/go-uuid"
	"io/ioutil"
	"log"
	"net/http"
)

var southUrl = flag.String("southUrl", "http://admin.skyworthtest.top/api/skyworthSouthbound/mqttIpConfig/getClusterIp?productKey=bench_mqtt", "南向地址获得mqttconfig 地址")

/**
从南向获得分配到的mqtt节点
*/
func GetMqttClusterIp(southUrl string) string {
	uuidStr, _ := uuid.GenerateUUID()
	resp, err := http.Get(southUrl + "&deviceKey=" + uuidStr)
	if err != nil {
		log.Print("访问中心节点失败", err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	str2 := string(body)
	fmt.Println(str2)
	json, _ := simplejson.NewJson(body)
	code, _ := json.Get("code").Int()
	log.Print(code)
	var res string
	if code == 200 {
		res, _ = json.Get("data").Get("ip").String()
	} else {
		res, _ = json.Get("data").String()
	}
	log.Print(res)
	return res
}

