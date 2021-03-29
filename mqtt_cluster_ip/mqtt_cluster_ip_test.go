package mqtt_cluster_ip

import (
	"github.com/hashicorp/go-uuid"
	"testing"
)

func TestGetMqttClusterIp(t *testing.T) {
	//for i := 0; i < 600; i++ {
	//	for j := 0; j < 200; j++ {
	//		go GetMqttClusterIp()
	//	}
	//	log.Print("----------------------")
	//}
	//select {}
	uuidStr, _ := uuid.GenerateUUID()
	GetMqttClusterIp("", uuidStr)
}
