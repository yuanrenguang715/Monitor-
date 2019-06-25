# 监控脚本使用手册
      监控端口访问并发连接数&端口流量走向
      可使用Grafana展示
      InfluxDB数据层写入
# 可把本脚本写入Telegraf
## Grafana InfluxDB Telegraf
      InfluxDB配置：https://www.cnblogs.com/jackyroc/p/7677508.html
      Grafana配置：https://grafana.com/
      Telegraf配置：https://www.influxdata.com/time-series-platform/telegraf/
## 参数说明

-p,-db-url,-db是必须的参数,其它的参数都是可以省略的,都有默认(default)值

```text
Usage of ./monitor:
  -p string
        -p 需要监控的端口号,可以是多个用,分开,如： -p 8080,4040
  -db-url string
        -db-url influxdb提供的api路径,如：-db-url http://43.249.195.26:8086 这是influxdb提供的写库api
  -db
        -db 指定数据库 如：-db test
  -s
        -s 时间间隔,多长时间获取一次监控数据,单位：秒/s,如： -s 5

示列： 要给root权限
sudo ./monitor -p 1080 -db-url http://43.249.195.26:8086 -db dmonitor -s 5



 
