package utils

import (
	"fmt"
	"net"
	"strings"
)

//为了让读取配置文件的时候可以根据自己的ip地址去获取相应的key
//collect_log_key这个key的值
func GetOutboundIp() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()
	//其实并没有真正的分发请求
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.IP.String(), ":")[0]
	return
}
