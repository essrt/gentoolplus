# gentoolplus
Install（下载）：
go get -u github.com/essrt/gentoolplus@latest
go install github.com/essrt/gentoolplus

Useage（使用）：
gentoolplus -dbName dbname -dsn "user:pwd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

参数选项的解释：

	-dbName 	数据库名称（*必填）
	-dsn   		用于连接数据库的DSN
	-outPath	指定输出目录(默认 ./dao/query)
	-c 		配置文件路径(默认 ./gentoolplus_config.json)，命令行选项的优先级高于配置文件
 	-h 		帮助文档

配置文件示例:

	{
	    "version": "0.1",
	    "database": {
	        "dbDriver": "mysql",
	        "dbName": "dbname",
	        "dsn": "root:pwd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
	        "outPath": "./dao/query",
	        "dataMap": {
	            "tinyint": "int64",
	            "smallint": "int64",
	            "mediumint": "int64",
	            "bigint": "int64",
	            "int": "int64"
	        },
	        "fieldNullable": true,
	        "fieldCoverable": false,
	        "fieldSignable": false,
	        "fieldWithIndexTag": false,
	        "fieldWithTypeTag": true,
	        "withUnitTest": false
	    }
	}

详细配置文件参数说明：

	dbDriver		string			数据库引擎(目前只支持MySQL，默认值：mysql)
 	dbName  		string          	数据库名称  
	dsn     		string          	用于连接数据库的DSN  
	outPath 		string          	指定输出目录(默认值：./dao/query) 
	dataMap 		map[string]string   	数据库自定义字段的数据类型
	fieldNullable 		bool   			表字段可为 null 值时, 对应结体字段使用指针类型，默认值：false
	fieldCoverable 		bool 			当字段具有默认值时生成指针，以解决无法分配零值的问题，默认值：false
	fieldSignable 		bool  			模型结构体字段的数字类型的符号表示是否与表字段的一致, false指示都用有符号类型，默认值：false
	fieldWithIndexTag 	bool  			生成 gorm 标签的字段索引属性，默认值：false
	fieldWithTypeTag 	bool 			生成 gorm 标签的字段类型属性，默认值：false
	withUnitTest 		bool  			生成单元测试，默认值：false,
	outFile 		string    		Genrated 查询代码文件名称，默认值：gen.go
	modelPkgPath 		string   		生成模型代码包名称。默认值：model  
