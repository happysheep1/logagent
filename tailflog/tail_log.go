package tailflog

import (
	"fmt"
	"github.com/hpcloud/tail"
	"golang.org/x/net/context"
	"test.com/studygo/logagent/kafka"
)

//一个日志收集的任务,这时就可以使用多个tailobj
type TailTask struct {
	path       string
	topic      string
	instance   *tail.Tail
	ctx        context.Context
	cancelFunc context.CancelFunc //为了实现可以退出某个监听
}

func NewTailTask(path, topic string) (tailObj *TailTask) {
	ctx, cancel := context.WithCancel(context.Background())
	tailObj = &TailTask{
		path:       path,
		topic:      topic,
		ctx:        ctx,
		cancelFunc: cancel,
	}
	tailObj.init() //根据路径打开对应的日志信息
	return
}
func (t TailTask) init() {
	config := tail.Config{
		ReOpen:   true, //重新打开
		Follow:   true, //是否跟随
		Location: &tail.SeekInfo{Offset: 0, Whence: 2},
		//从文件的哪个地方开始读
		MustExist: false, //不存在就报错
		Poll:      true,  //

	}
	var err error
	t.instance, err = tail.TailFile(t.path, config)

	if err != nil {
		fmt.Println("tail file failed, err:", err)
		return
	}
	//那什么时候实现退出呢
	//goroutine执行的函数退出的时候就退出
	//手动告诉他该退出
	go t.run() //开始采集日志发送到kafka

}

func (t *TailTask) ReadLog() <-chan *tail.Line { //返回一个只读的chan
	return t.instance.Lines

}

//具体收集日志的方法
func (t *TailTask) run() {
	for {
		select {
		case <-t.ctx.Done(): //这里就是在实现这个goroutine退出
			fmt.Printf("tailtask %s_%s退出啦", t.path, t.topic)
			return
		case line := <-t.instance.Lines:
			//函数调用函数，这时候是同步到的
			//kafka.Send2Kafka(t.topic, line.Text)
			//	考虑需要进行改进---->管道

			//	改进方式
			//先把日志数据发送到一个管道中
			kafka.Send2Chan(t.topic, line.Text)
			//kafka中有单独的goroutine去取日志数据发送到kafka

		}
	}
}
