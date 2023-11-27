# gentoolplus
Install（下载）：
go get github.com/essrt/gentoolplus
go install github.com/essrt/gentoolplus

Useage（使用）：
gentoolplus  -dsn "user:pwd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local" -dbName dbname

参数选项的解释：

	-dbName: 数据库名称（必须提供一个数据库名称）
	-dsn "user:pwd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"  (用于连接数据库的DSN)
	-outPath：指定输出目录(默认 “./dao/query”)
	-c 配置文件路径、默认值 “”、命令行选项的优先级高于配置文件

配置文件格式为json格式
主要包含：

 	dbName  string          数据库名称  
	dsn     string          用于连接数据库的DSN  
	outPath string          指定输出目录(默认 “./dao/query”) 
	dataMap map[string]string   (自定义字段的数据类型)
	fieldNullable bool   表字段可为 null 值时, 对应结体字段使用指针类型，默认为false
	fieldCoverable bool 当字段具有默认值时生成指针，以解决无法分配零值的问题，默认为false
	fieldSignable bool  模型结构体字段的数字类型的符号表示是否与表字段的一致, false指示都用有符号类型，默认为false
	fieldWithIndexTag bool  生成 gorm 标签的字段索引属性，默认为false
	fieldWithTypeTag bool 生成 gorm 标签的字段类型属性，默认为false
	withUnitTest bool  生成单元测试，选项: false / true，默认值 false,
	OutFile string    Genrated 查询代码文件名称，默认值：gen.go
	ModelPkgPath string   生成模型代码包名称。默认值：model  
