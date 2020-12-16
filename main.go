package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"sync"
	"test.com/studygo/logagent/config"
	"test.com/studygo/logagent/etcd"
	"test.com/studygo/logagent/kafka"
	"test.com/studygo/logagent/tailflog"
	"test.com/studygo/logagent/utils"
	"time"
)

var (
	c  = new(config.AppConfig)
	wg sync.WaitGroup
)

//func run() {
//
//	//1.读取日志
//	for {
//		select {
//		case line := <-tailflog.ReadLog():
//			//	2.发送到kafka
//			kafka.Send2Kafka(c.KafkaConf.Topic, line.Text)
//		default:
//			time.Sleep(time.Second)
//		}
//	}
//
//}

//logagent入口文件
func main() {

	err := ini.MapTo(c, "./config/config.ini")
	if err != nil {
		fmt.Printf("load config.ini error err:%v", err)
		return
	}

	//	1.初始化kafka链接
	err = kafka.Init([]string{c.KafkaConf.Address}, c.KafkaConf.ChanMaxSize)
	if err != nil {
		fmt.Printf("init kafka error err:%v", err)
		return
	}
	fmt.Println("init kafka success")
	//	2.打开日志文件准备收集
	//	err=tailflog.Init(c.TailflogConf.Path)
	//	if err != nil {
	//		fmt.Printf("init tailflog error err:%v",err)
	//		return
	//	}
	//	fmt.Println("init tailflog success")
	//	run()

	//初始化etcd
	err = etcd.Init(c.EtcdConfig.Address, time.Duration(c.EtcdConfig.Timeout)*time.Second)
	if err != nil {
		fmt.Printf("init etcd error err:%v", err)
		return
	}
	fmt.Println("init etcd success")

	//	2.1从etcd中获取日志收集项的配置信息
	//为了实现每个logagent拉取自己的配置，以ip地址作为区分的依据
	ipStr, err := utils.GetOutboundIp()
	if err != nil {
		fmt.Printf("获取本地ip地址失败 err:%v", err)
		return
	}
	EtcdConfigCollectLogKey := fmt.Sprintf(c.EtcdConfig.CollectLogKey, ipStr)
	logEntryConf, err := etcd.GetConf(EtcdConfigCollectLogKey)
	if err != nil {
		fmt.Printf("etcd.GetConf failed,err:%v\n", err)
		return
	}
	//2.2派一个哨兵监视日志收集项的变化（有变化及时通知我的logagent实现热加载配置）

	fmt.Printf("get conf from etcd success, %v\n", logEntryConf)
	for index, value := range logEntryConf {
		fmt.Printf("index:%v value:%v\n", index, value)
	}
	//	3.收集日志发往kafka
	//3.1循环每一个收集项，创建tailfObj
	//3.2发往kafaka
	tailflog.Init(logEntryConf)
	wg.Add(1)
	//新增监听etcd配置项更改时的通知
	//后台一直运行且不返回
	//tailflog.PushNewConf()返回的是一个只写的存放新通道的函数
	go etcd.WatchConf(EtcdConfigCollectLogKey, tailflog.PushNewConf())
	//WatchConf()发现新配置就往这个通道里面写值
	//tailflog.updateConf()会知道这个新增的值进行更改
	wg.Wait() //不会走到这一步，因为上面的函数一直在for循环

	//
	//
	//for _,conf:=range logEntryConf{
	////	conf是一个类似index:0 value:&{e:/tmp/nginx.log web_log}

	//低效的写法，这是为什么需要TailTask的原因

	//	config := tail.Config{
	//		ReOpen:    true,  //重新打开
	//		Follow:    true,  //是否跟随
	//		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
	//		//从文件的哪个地方开始读
	//		MustExist: false, //不存在就报错
	//		Poll: true, //
	//
	//	}
	//	tailObj, err := tail.TailFile(conf.Path, config)
	//	if err != nil {
	//		fmt.Println("tail file failed, err:", err)
	//		return
	//	}
	//	for {
	//		select {
	//		case line:=<-tailObj.Lines:
	//			kafka.Send2Kafka(conf.Topic,line.Text)
	//
	//		}
	//	}
	//}

}
