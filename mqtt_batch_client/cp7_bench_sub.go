package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hashicorp/go-uuid"
	"gopkg.in/robfig/cron.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"mqtt-batch-client-go/local_ip"
	"mqtt-batch-client-go/mapstring"
	"mqtt-batch-client-go/redis"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

//var urlString = flag.String("url", "mqtts://124.70.73.148:8883", "broker url")
var urlString = flag.String("url", "mqtt://114.116.244.139:1883", "broker url")
var workers = flag.Int("workers", 1000, "number of workers")
var qos = flag.Uint("qos", 1, "sub qos level")
var clearsession = flag.Bool("clear", true, "clear session")

var ca = flag.String("ca", "./root-cacert.pem", "ca")
var clientCertFile = flag.String("clientCertFile", "C:\\Users\\sdt15557\\GolandProjects\\coolpy7_benchmark\\test_bench_sub\\client-cert.pem", "clientCertFile")
var clientKeyFile = flag.String("clientKeyFile", "C:\\Users\\sdt15557\\GolandProjects\\coolpy7_benchmark\\test_bench_sub\\client.key", "clientKeyFile")

var appName = flag.String("appName", "test", "appName")
var developerId = flag.String("developerId", "90b9aa7e25f80cf4f64e990b78a9fc5ebd6cecad", "developerId")
var pk = flag.String("pk", "bench_mqtt", "productKey")

// 初始化 push
var initPushFlag = flag.Bool("initPushFlag", true, "init push flag")

// 并发数
var currNum = flag.Int("currNum", 1000, "并发数")

// 间隔
var gap = flag.Int("gap", 200, "间隔ms数")

// cron
var cronStr = flag.String("cron", "*/1 * * * * *", "cron表达式, 默认每秒")

// 发送消息数量
var publishNum = flag.Int("publishNum", 1000, "发送消息的协程数量")

var randomPush = flag.Int("randomPush", 0, "cron后随机延时发送时间,单位min")

var redisAddress = flag.String("redisAddress", "", "redis连接地址")
var redisAuth = flag.String("redisAuth", "", "redis密码")
var setNodeNo = flag.Int("setNodeNo", -1, "指定节点编号")

var conNum int32
var currentNode int

var currentNodes = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}

func init() {
	if *setNodeNo != -1 {
		currentNode = *setNodeNo
		return
	}
X:
	for x, nodeNo := range currentNodes {
		//log.Printf("nodeNo %d", nodeNo)
		lock := redis.NewRedisPool(*redisAddress, *redisAuth)
		if lock.AddLockDisExpire("batch_"+strconv.Itoa(nodeNo), ip.String()) {
			currentNode = nodeNo
			break
		} else {
			ipGet := lock.GetLock("batch_" + strconv.Itoa(nodeNo))
			if ipGet == ip.String() {
				currentNode = nodeNo
				break
			}
			log.Printf("未找到NO ！！！now: %d", x)
			if nodeNo == 21 {
				goto X
			}
		}
	}
	log.Printf("根据IP %s 从redis 分布式锁获取 node %d", ip.String(), currentNode)
}

var ip, err = local_ip.ExternalIP()

func connAndSub() {
	log.Print("参数初始化")
	num := *workers / *currNum
	if num == 0 {
		num = 1
	}
	// 获得当前Ip
	if err != nil {
		log.Println(err)
	}
	log.Println(ip.String())
	clientTopic := "/client/" + ip.String()
	log.Printf("当前客户端 center 通信 clientTopic： %s", clientTopic)
	fmt.Printf("%d 并发测试连接开始``````````````` ", *workers)
	for i := 0; i < num; i++ {
		for j := 0; j < *currNum; j++ {
			go createConn()
		}
		time.Sleep(time.Duration(*gap) * time.Millisecond)
	}
	fmt.Printf("%d 并发测试连接开始``````````````` ", *workers)
	fmt.Println("finish", workers)
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range signalChan {
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}

func createConn() {
	atomic.AddInt32(&conNum, 1)
	loadConNum := atomic.LoadInt32(&conNum)
	i := int(loadConNum)
	t := time.Now()
	fmt.Printf("序号：[%d] 开始计时\n", i)
	id := strconv.Itoa(i)
	uuid, _ := uuid.GenerateUUID()
	dk := strconv.Itoa(currentNode) + "test" + id
	urlArr := strings.Split(*urlString, ",")
	num := i % len(urlArr)
	url := urlArr[num]
	clientId := dk + "_" + uuid
	opts := mqtt.NewClientOptions().AddBroker(url).SetClientID(clientId)
	//自动重连机制，如网络不稳定可开启
	opts.SetAutoReconnect(true)      //启用自动重连功能
	opts.SetMaxReconnectInterval(10) //每30秒尝试重连
	opts.SetConnectRetry(true)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(60 * time.Second)
	// 设置消息回调处理函数
	opts.SetDefaultPublishHandler(f)
	opts.SetCleanSession(*clearsession)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Println("OnConnectionLost!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", clientId, err.Error())
		log.Println("OnConnectionLost!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", err.Error())
	})
	opts.SetUsername("skyworth")
	opts.SetPassword("sky@123!")

	tlsconfig := NewTLSConfig()
	opts.SetTLSConfig(tlsconfig)

	c := mqtt.NewClient(opts)
RECONNECT:
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		//panic(token.Error())
		log.Println(token.Error())
		log.Println("重连")
		goto RECONNECT
	}

	var topicArr []string
	broadcastTopic := fmt.Sprintf("/sdt/down/push/broadcast/%v/%v", *appName, *pk)
	dkTopic := fmt.Sprintf("/sdt/down/push/%v/%v/%v", *appName, *pk, dk)
	aliasTopic := fmt.Sprintf("/sdt/down/push/%v/%v/%v/%v", *appName, *pk, "deviceId", dk)

	iotCommand := "/sdt/down/iot/data/command"
	topicArr = append(topicArr, broadcastTopic, iotCommand, "testest", dkTopic, aliasTopic)

	for _, topic := range topicArr {
		if token := c.Subscribe(topic, byte(1), nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
		}
	}

	cr := cron.New()
	if *initPushFlag && *publishNum != 0 {
		cronArr := strings.Split(*cronStr, ",")
		for corniter := range cronArr {
			_, _ = cr.AddFunc(cronArr[corniter], func() {
				//eventMap := map[string]interface{}{"topic":"/sdt/up/iot/data/event","tenantId":*developerId,"productKey":*pk,"messageId":"11112","appName":*appName,"userId":dk,"gateway":"0","action":"event","event_name":"properties","event_type":"info","image_inversion":"1","test":"1","service_id":"setting"}
				eventMap := map[string]interface{}{"topic": "/sdt/up/iot/data/event", "tenantId": *developerId, "productKey": *pk, "messageId": "11112", "appName": *appName, "userId": dk, "data": "{\"gateway\":\"0\",\"action\":\"event\",\"event_name\":\"properties\",\"event_type\":\"info\",\"image_inversion\":\"1\",\"test\":\"1\",\"service_id\":\"setting\"}"}
				//property := map[string]interface{}{"topic":"/sdt/up/iot/data/property","tenantId":*developerId,"productKey":*pk,"messageId":"11112","appName":*appName,"userId":dk,"data":"{\"gateway\":\"0\",\"action\":\"property\",\"image_inversion\":\"1\",\"test\":\"1\",\"service_id\":\"setting\"}"}
				//c.Publish("/sdt/up/iot/data/property", byte(1), false, mapstring.MapToJson(property))
				if *randomPush != 0 {
					num := rand.Intn(*randomPush)
					time.Sleep(time.Duration(num) * time.Second)
				}
				if token := c.Publish("/sdt/up/iot/data/event", byte(1), false, mapstring.MapToJson(eventMap)); token.Wait() && token.Error() != nil {
					log.Printf("message publish error %d", token.Error())
				} else {
					log.Printf("message publish success")
				}
			})
		}
		*publishNum--
	}
	cr.Start()
	//defer cr.Stop()
	elapsed := time.Since(t)
	fmt.Printf("序号[%d] DK[%s] app elapsed: %s \n", loadConNum, dk, elapsed)
	if i%1000 == 0 {
		log.Printf("已经连接连接数 %d", loadConNum)
	}
}

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	log.Printf("client 收到消息: %s", msg)
}

func NewTLSConfig() *tls.Config {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(*ca)
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(*clientCertFile, *clientKeyFile)
	if err != nil {
		panic(err)
	}

	// Just to print out the client certificate..
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}
	//fmt.Println(cert.Leaf)

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}
}

func disconnectHandler() {

}

func main() {
	flag.Parse()
	fmt.Println("从中心节点获得mqtt集群地址:", urlString)
	fmt.Println("workers:", *workers)
	fmt.Println("clearsession:", *clearsession)
	fmt.Println("QOS:", *qos)
	fmt.Println("developerId:", *developerId)
	fmt.Println("appName:", *appName)
	fmt.Println("pk:", *pk)
	fmt.Println("ca地址:", *ca)
	fmt.Println("clientCertFile地址:", *clientCertFile)
	fmt.Println("clientKeyFile地址:", *clientKeyFile)
	connAndSub()

}
