## 日志配置
```yaml
log:                                      #所有日志配置
  default:                                  #默认日志配置，log.Debug("xxx")
    - writer: console                         #控制台标准输出 默认
      level: debug                            #标准输出日志的级别
    - writer: file                              #本地文件日志
      level: info                               #本地文件滚动日志的级别
      formatter: json                           #标准输出日志的格式
      writer_config:                            #本地文件输出具体配置
        filename: ../log/quote.log          #本地文件滚动日志存放的路径
        roll_type: size                         #文件滚动类型,size为按大小滚动
        max_age: 7                              #最大日志保留天数
        max_size: 10                            #本地文件滚动日志的大小 单位 MB
        max_backups: 10                         #最大日志文件数
        compress:  false                        #日志文件是否压缩
    - writer: file                              #本地文件日志
      level: info                               #本地文件滚动日志的级别
      formatter: json                           #标准输出日志的格式
      writer_config:                            #本地文件输出具体配置
        filename: ../log/time.log          #本地文件滚动日志存放的路径
        roll_type: time                         #文件滚动类型,time为按时间滚动
        max_age: 7                              #最大日志保留天数
        time_unit: day                          #滚动时间间隔，支持：minute/hour/day/month/year
    - writer: atta                                #atta远程日志输出
      remote_config:                              #远程日志配置，业务自定义结构，每一种远程日志都有自己独立的配置
        atta_id: '05e00006180'                    #atta id 每个业务自己申请
        atta_token: '6851146865'                  #atta token 业务自己申请
        message_key: msg                          #日志打印包体的对应atta的field
        field:                                    #申请atta id时，业务自己定义的表结构字段，顺序必须一致
          - msg
          - uid
          - cmd
  custom:                                   #业务自定义的logger配置，名字随便定，每个服务可以有多个logger，可使用 log.Get("custom").Debug("xxx") 打日志
    - writer: file                              #业务自定义的core配置，名字随便定
      level: info                               #业务自定义core输出的级别
      writer_config:                            #本地文件输出具体配置
        filename: ../log/quote1.log               #本地文件滚动日志存放的路径
    - writer: file                              #本地文件日志
      level: info                               #本地文件滚动日志的级别
      writer_config:                            #本地文件输出具体配置
        filename: ../log/quote2.log               #本地文件滚动日志存放的路径
```
