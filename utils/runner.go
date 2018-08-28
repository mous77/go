package utils

import (
	"fmt"
	"sync/atomic"
	"bytes"
	"errors"
	"time"
	"sync"
	"lib/log"
)

const (
	tsWait int32 = 0 // 任务 等待执行
	tsBusy int32 = 1 // 任务 正在执行
	tsTerm int32 = 2 // 任务 执行结束
)

// 任务调度器 类型
type tRunner struct {
	lg log.ILogger // 日志对象

	taskSeq  uint64             // 任务流水号
	taskMaps map[uint64]*tRTask // 任务哈希表
	taskReqs chan *tRTask       // 请求的任务队列
	taskWait chan *tRTask       // 排队的任务队列

	threadMap  map[int32]*tThread // 线程表
	threadSeq  int32              // 线程流水号
	threadBusy int32              // 忙的线程数
	threadFree int32              // 空闲的线程数

	once   *sync.Once // 初始化标识
	active int32      // 活动标识
}

var runner = &tRunner{once: &sync.Once{}}

func getRunner() (*tRunner) {
	runner.once.Do(func() {
		// 这个动作只要做一次即可
		runner.lg = log.GetLogger("runner")
		runner.active = 1

		runner.threadMap = make(map[int32]*tThread)

		runner.taskMaps = make(map[uint64]*tRTask)

		runner.taskWait = make(chan *tRTask, 1024)
		runner.taskReqs = make(chan *tRTask, 1024)

		go runner.loopRequest() // 调度者
		go runner.loopStatus()  // 报告者
	})

	return runner
}

func (r *tRunner) isActive() (bool) {
	return 1 == atomic.LoadInt32(&r.active)
}

func (r *tRunner) loopStatus() {
	for r.isActive() {
		r.lg.Info("[thread(map:%d,free:%d, busy:%d), task(map:%d, req:%d, wait:%d)]",
			len(r.threadMap), r.threadFree, r.threadBusy, len(r.taskMaps), len(r.taskReqs), len(r.taskWait))
		time.Sleep(time.Second * 5)
	}
}

//
// 在循环体内，对请求的队列进行处理：
// 如果任务状态是待执行，则放入等待队列
// 如果任务状态是已结束，则将任务删除
func (r *tRunner) loopRequest() {

	for t := range r.taskReqs {
		r.lg.Trace("req=%s", t.String())

		// 根据任务状态，添加或删除 任务
		switch t.sts {
		case tsWait:
			if !r.isActive() {
				return
			}

			r.taskMaps[t.seq] = t
			r.taskWait <- t

			// 加了任务，看是否有足够的线程来执行，不够就加
			if int32(len(r.taskWait)) > atomic.LoadInt32(&r.threadFree) {
				r.newThread()
			}

		case tsTerm:
			delete(r.taskMaps, t.seq)
		}
	}
}

func (r *tRunner) newTask(name string, sts int32) (*tRTask) {
	if nil == r.lg {
		panic(errors.New("setup first!"))
	}

	if atomic.LoadInt32(&r.active) == 0 {
		msg := fmt.Sprintf("can not launch (%s) on close runner", name)
		panic(errors.New(msg))
	}

	// gen new task
	seq := atomic.AddUint64(&r.taskSeq, 1)
	key := fmt.Sprintf("[%02X](%s)", seq, name)

	//r.lg.Debug("newTask(%s)", key)

	// 生成性的任务对象，并加入请求队列
	t := &tRTask{
		seq: seq,
		key: key,
		sts: sts,
	}

	return t
}

// 加入一个新任务，具体步骤如下
// 1、检查执行器是否活动
// 2、生成性任务对象，设置状态为 待执行
// 3、将任务放入请求队列
func (r *tRunner) execute(fun func(interface{}), name string, arg interface{}) {
	t := r.newTask(name, tsWait)

	t.fun = fun
	t.arg = arg

	if r.isActive() {
		r.taskReqs <- t
	}
}

/**
	get the names of the task list
 */
func (r *tRunner) GetTasks() (string) {
	buf := bytes.NewBufferString(fmt.Sprintf("tasks:%d=\this\n", len(r.taskMaps)))
	for i, n := range r.taskMaps {
		buf.WriteString(fmt.Sprintf("\this%d=%s\this\n", i, n))
	}
	return buf.String()
}

func (r *tRunner) shutdown() {
	if atomic.CompareAndSwapInt32(&r.active, 1, 0) {
		r.lg.Info("total worker=%d", len(r.threadMap))
		close(r.taskReqs)
		close(r.taskWait)
	}
}

func (r *tRunner) newThread() {
	// 新的线程开始，活动线程数加1
	atomic.AddInt32(&r.threadFree, 1)

	tid := atomic.AddInt32(&r.threadSeq, 1)
	threadObj := &tThread{threadID: tid, tasks: r.taskWait}
	r.threadMap[tid] = threadObj

	r.lg.Debug("newThread[%d]", tid)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go threadObj.run(wg)
	wg.Wait()
}

// 给外部调用的接口
// 对内直接调用 调度器的 launch 方法即可
func GoRunAdd(_fun func(_arg interface{}), _name string, _arg ... interface{}) {
	var arg interface{} = nil
	if len(_arg) > 0 {
		arg = _arg[0]
	}

	getRunner().execute(_fun, _name, arg)
}

func GoExecute(f func(), name string) {
	GoRunAdd(func(interface{}) { f() }, name)
}

// 终止任务调度器
func GoRunTerm() {
	getRunner().shutdown()
}

// 线程对象
type tThread struct {
	threadID int32
	tasks    <-chan *tRTask
}

// 线程执行者动作是： 不断从等待队列中，取出待执行的任务，随后执行
func (th *tThread) run(_wg *sync.WaitGroup) {
	_wg.Done()

	for task := range th.tasks {
		//runner.lg.Debug("switch-to-thread[%d] %s", this.threadID, task.key)
		task.exec()
	}
}

// 任务对象
type tRTask struct {
	seq uint64 // 任务流水号
	key string // 任务关键字
	sts int32  // 任务状态 0:待执行, 1:已结束

	fun func(interface{}) // 任务方法
	arg interface{}       // 任务参数
}

func (t *tRTask) String() (string) {
	return fmt.Sprintf("task[%d](%s)", t.seq, t.key)
}

// 执行一个任务
func (t *tRTask) exec() {
	//runner.lg.Debug("exec(%s)", this.key)
	defer func() {
		if v := recover(); nil != v {
			runner.lg.Error("error on exec %s %v", t.String(), v)
		}

		// 任务执行完成后，将其标识设置为 结束
		atomic.StoreInt32(&t.sts, tsTerm)

		// 同时将该任务对象，让如到请求队列中，便于从活动队列删除
		if runner.isActive() {
			runner.taskReqs <- t
		}

		atomic.AddInt32(&runner.threadBusy, -1)
		atomic.AddInt32(&runner.threadFree, 1)
		//runner.lg.Debug("done %s", this.key)
	}()

	atomic.AddInt32(&runner.threadFree, -1)
	atomic.AddInt32(&runner.threadBusy, 1)

	t.fun(t.arg)
}

func TestRunner() {
	log.Setup(&log.TConfig{Level: log.LLFINE, Console: true})
	getRunner()
	for i := 0; i < 100; i++ {
		GoRunAdd(func(_arg interface{}) {
			runner.lg.Info("----- hello %v", i)
			time.Sleep(time.Second)
		}, fmt.Sprintf("task-%d", i), i)
	}
	time.Sleep(time.Minute)
	runner.shutdown()
}
