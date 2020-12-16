package tailflog

import (
	"fmt"
	"test.com/studygo/logagent/etcd"
	"time"
)

var taskMgr *tailfLogMgr //定义一个全局管理者

//保存现在所有的日志采集信息保存起来
type tailfLogMgr struct {
	logEntry    []*etcd.LogEntry //就是记录了path和topic的json字段
	taskMap     map[string]*TailTask
	newConfChan chan []*etcd.LogEntry //新的配置

}

func Init(logEntry []*etcd.LogEntry) {
	taskMgr = &tailfLogMgr{
		logEntry:    logEntry,
		taskMap:     make(map[string]*TailTask, 16),
		newConfChan: make(chan []*etcd.LogEntry), //无缓冲区的通道
	}
	for _, conf := range taskMgr.logEntry {
		//	conf是一个类似index:0 value:&{e:/tmp/nginx.log web_log}

		//初始化的时候起了多少个task得记起来
		//后续用来判断新增删除了什么
		tailtask := NewTailTask(conf.Path, conf.Topic)
		//taskMgr.taskMap[conf.Path]=tailtask这样不好，因为可能是值改变
		mk := fmt.Sprintf("%s_%s", conf.Path, conf.Topic)
		taskMgr.taskMap[mk] = tailtask

	}
	go taskMgr.updateConf()
}

//当有新的配置时进行的更新操作
//配置新增
//配置删除
//配置变更
func (t *tailfLogMgr) updateConf() {
	for {
		select {
		case newConf := <-t.newConfChan: //没有值的时候一直阻塞
			for _, conf := range newConf {
				//	判读conf.path在不在map里面
				mk := fmt.Sprintf("%s_%s", conf.Path, conf.Topic)
				if _, ok := t.taskMap[mk]; ok {
					//没有更改过的
					continue
				} else {
					//新增和更改都认为是新增的统一处理，
					//最后再删掉logentry有但是newconf中没有的就可以了
					tailObj := NewTailTask(conf.Path, conf.Topic)
					taskMgr.taskMap[mk] = tailObj
				}
				//删掉logentry有但是newconf中没有的
				for _, c1 := range t.logEntry {
					isdelete := true
					for _, c2 := range newConf {
						if c1.Path == c2.Path && c1.Topic == c2.Topic {
							isdelete = false
							continue
						}
						if isdelete {
							//	把c1对应的tailobj删掉
							//	那怎么停呢
							mk := fmt.Sprintf("%s_%s", c1.Path, c1.Topic)
							t.taskMap[mk].cancelFunc() //这时就会发送ctx.done()

						}
					}
				}

			}
			//配置新增
			//配置删除
			//配置变更
			fmt.Println("新的配置", newConf)
		default:
			time.Sleep(time.Second)

		}
	}
}

//向外暴露一个newConfChan
func PushNewConf() chan<- []*etcd.LogEntry {
	return taskMgr.newConfChan
}
