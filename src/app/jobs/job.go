package jobs

import (
	//"bytes"
	//"os/exec"
	//"app/mail"
	//"html/template"
	
	"fmt"
	"errors"
	"io/ioutil"	
	"app/models"
	//"runtime/debug"
	"strings"
	"time"
	"net/http"
	"app/redis"
	//"github.com/axgle/mahonia"
	"github.com/imroc/req"
	"github.com/astaxie/beego"
)

type Job struct {
	id         int                                               // 任务ID
	logId      int64                                             // 日志记录ID
	name       string                                            // 任务名称
	task       *models.Task                                      // 任务对象
	runFunc    func(time.Duration) (string, string, error, bool) // 执行函数
	status     int                                               // 任务状态，大于0表示正在执行中
	Concurrent bool                                              // 同一个任务是否允许并行执行
}

func NewJobFromTask(task *models.Task) (*Job, error) {
	if task.Id < 1 {
		return nil, fmt.Errorf("ToJob: 缺少id")
	}
	job := NewCommandJob(task)
	job.task = task
	job.Concurrent = task.Concurrent == 1
	return job, nil
}

func NewCommandJob(task *models.Task) *Job {
	job := &Job{
		id:   task.Id,
		name: task.TaskName,
	}
	job.runFunc = func(timeout time.Duration) (string, string, error, bool) {
		header := make(http.Header)
		if task.ApiHeader != "" && strings.TrimSpace(task.ApiHeader) != "" {
			headers := strings.Split(task.ApiHeader, "\n")
			for _,val := range headers {
				keyval := strings.Split(val, "=")
				if len(keyval) > 0 {
					v := strings.TrimSpace(keyval[0])
					v1 := strings.TrimSpace(keyval[1])
					if v != "" && v1 != "" {
						header.Set(v, v1)
					} else {
						continue
					}
				}
			}
		}	
		//fmt.Println(header)
		
		responsestr := ""
		var err error
		var res *req.Resp
		
		//这里还没有处理超时
		if task.ApiMethod == "POST" {
			if task.PostBody != "" {
				contenttype := header.Get("Content-Type")			
				//如果没有设置就用json方式提交
				if contenttype == "" || contenttype == "application/json" {
					res, err = req.Post(task.ApiUrl, header, req.BodyJSON(task.PostBody))
				} else {
					res, err = req.Post(task.ApiUrl, header, req.BodyXML(task.PostBody))
				}
			} else {
				res, err = req.Post(task.ApiUrl, header)
			}
			
		} else {
			res, err = req.Get(task.ApiUrl, header)
		}
		
		if err == nil {
			bodystr, _ := ioutil.ReadAll(res.Response().Body)
			defer res.Response().Body.Close()

			responsestr = string(bodystr)
			//fmt.Println(responsestr)
			//encoder := mahonia.NewDecoder("gbk")
			
			if res.Response().StatusCode != 200 {
				//return encoder.ConvertString(responsestr), "", errors.New(fmt.Sprintf("返回的状态码为：%s", res.Response().StatusCode)), false
				return responsestr, "", errors.New(fmt.Sprintf("返回的状态码为：%s", res.Response().StatusCode)), false
			}
			
			//return encoder.ConvertString(responsestr), "", nil, false
			return responsestr, "", nil, false
		} else {
			return "", "", err, false
		}
	}
	return job
}

func (j *Job) Status() int {
	return j.status
}

func (j *Job) GetName() string {
	return j.name
}

func (j *Job) GetId() int {
	return j.id
}

func (j *Job) GetLogId() int64 {
	return j.logId
}

func (j *Job) Run() {
	if !j.Concurrent && j.status > 0 {
		beego.Warn(fmt.Sprintf("任务[%d]上一次执行尚未结束，本次被忽略。\n", j.id))
		return
	}

	defer func() {
		if err := recover(); err != nil {
			beego.Error(err, "\n")
		}
	}()

	if workPool != nil {
		workPool <- true
		defer func() {
			<-workPool
		}()
	}

	beego.Debug(fmt.Sprintf("开始执行任务: %d\n", j.id))

	j.status++
	defer func() {
		j.status--
	}()

	t := time.Now()
	timeout := time.Duration(time.Hour * 24)
	if j.task.Timeout > 0 {
		timeout = time.Second * time.Duration(j.task.Timeout)
	}

	cmdOut, cmdErr, err, isTimeout := j.runFunc(timeout)

	ut := time.Now().Sub(t) / time.Millisecond

	// 插入日志
	log := new(models.TaskLog)
	log.TaskId = j.id
	log.Output = cmdOut
	log.Error = cmdErr
	log.ProcessTime = int(ut)
	log.CreateTime = t.Unix()
	
	var errredis error = nil
	//此时获取到了新数据，同步更新到redis
	if err == nil {
		errredis = redis.Set(j.task.CacheKey, cmdOut)
	}

	if isTimeout {
		log.Status = models.TASK_TIMEOUT
		log.Error = fmt.Sprintf("任务执行超过 %d 秒\n----------------------\n%s\n", int(timeout/time.Second), cmdErr)
	} else if err != nil {
		log.Status = models.TASK_ERROR
		log.Error = err.Error() + ":" + cmdErr
	} else if errredis != nil {
		log.Status = models.TASK_ERROR
		log.Error = "save redis has wrong: " + errredis.Error()
	}
	
	j.logId, _ = models.TaskLogAdd(log)

	// 更新上次执行时间
	j.task.PrevTime = t.Unix()
	j.task.ExecuteTimes++
	j.task.Update("PrevTime", "ExecuteTimes")
}
