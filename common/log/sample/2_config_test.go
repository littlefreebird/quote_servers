package sample

import (
	"quote/common/log"
	"testing"
)

const strConfigTest = `
  default:                                  #默认日志配置，log.Debug("xxx")
    - writer: console                         #控制台标准输出 默认
      level: debug                            #标准输出日志的级别
    - writer: file                              #本地文件日志
      level: info                               #本地文件滚动日志的级别
      formatter: json                           #标准输出日志的格式
      writer_config:                            #本地文件输出具体配置
        filename: ./go_play.log          #本地文件滚动日志存放的路径
        roll_type: size                         #文件滚动类型,size为按大小滚动
        max_age: 7                              #最大日志保留天数
        max_size: 10                            #本地文件滚动日志的大小 单位 MB
        max_backups: 10                         #最大日志文件数
        compress:  false                        #日志文件是否压缩
    - writer: file                              #本地文件日志
      level: info                               #本地文件滚动日志的级别
      formatter: json                           #标准输出日志的格式
      writer_config:                            #本地文件输出具体配置
        filename: ./time.log          #本地文件滚动日志存放的路径
        roll_type: time                         #文件滚动类型,time为按时间滚动
        max_age: 7                              #最大日志保留天数
        time_unit: day                          #滚动时间间隔，支持：minute/hour/day/month/year
  custom:                                   #业务自定义的logger配置，名字随便定，每个服务可以有多个logger，可使用 log.Get("custom").Debug("xxx") 打日志
    - writer: file                              #业务自定义的core配置，名字随便定
      level: info                               #业务自定义core输出的级别
      writer_config:                            #本地文件输出具体配置
        filename: ./go_play1.log               #本地文件滚动日志存放的路径
    - writer: file                              #本地文件日志
      level: info                               #本地文件滚动日志的级别
      writer_config:                            #本地文件输出具体配置
        filename: ./go_play2.log               #本地文件滚动日志存放的路径
`

func TestConfig(t *testing.T) {
	log.Setup([]byte(strConfigTest))
	log.Info("同时输出到控制台，go_play.log和time.log")
	log.Get("custom").Info("custom logger输出到两个文件")
	log.WithFields("国家", "新加坡").WithFields("祖籍", "马来西亚").Info("携带更多信息")
}
