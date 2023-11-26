# gentoolplus
Install（下载）：
go get github.com/essrt/gentoolplus
go install github.com/essrt/gentoolplus

Useage（使用）：
gentoolplus  -dbName dbname -dbPwd dbpwd -outPath ./query -dbUser root -dbHost localhost -dbPort 3306

-dbName: 数据库名称
-dbUser：数据库用户名称
-dbPwd: 连接数据库用户的密码
-dbHost: 数据库主机
-dbPort: 数据库的端口号
-outPath：指定输出目录(默认 “./query”)
