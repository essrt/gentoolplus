# gentoolplus
Install（下载）：
go get -u github.com/essrt/gentoolplus@latest
go install github.com/essrt/gentoolplus

Useage（使用）：
gentoolplus -dbName dbname -dsn "user:pwd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

参数选项的解释：
```
	-dbDriver	指定数据库引擎（mysql、postgres、sqlite、sqlserver），默认值：mysql
	-dbName 	数据库名称（*必填）
	-dsn   		用于连接数据库的DSN
	-outPath	指定输出目录(默认 ./dao/query)
	-outFile	指定输出文件(默认 gen.go)
	-c 		配置文件路径(默认 ./gentoolplus_config.json)，命令行选项的优先级高于配置文件
 	-h 		帮助文档
```
配置文件示例:
```
	{
    "version": "1.0",
    "database": {
        "dbDriver": "mysql",
        "dbName": "sqltest",
        "dsn": "user:pwd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        "outPath": "./dao/query",
        "outFile": "gentoolplus.go",
		"nspname": "public",
        "dataMap": {
            "tinyint": "int64",
            "smallint": "int64",
            "mediumint": "int64",
            "bigint": "int64",
            "int": "int64"
        },
        "fieldNullable": true,
        "fieldCoverable": true,
        "fieldSignable": false,
        "fieldWithIndexTag": false,
        "fieldWithTypeTag": false,
        "withUnitTest": false,
        "singularTable": true,
		"modelPkgPath": "model"
        "tables": [
			"address",
            "company",
			"department",
            "position",
            "staff",
            "staff_position",
			"workstation"
        ],
        "belongstoTables": {
			// 这里表示：staff belongs to company, staff belongs to department，所有跟staff有belongsto关系的表都写在这个数组中。
            "staff": [				
                "company", "department"
            ],
			...	// 其他表的belongsto关系					
        },
        "hasoneTables": {
			// 这里表示：staff has one address, staff has one workstation，所有跟staff有hasone关系的表都写在这个数组中。
			"staff": [				
				"address", "workstation"
			],
			...	// 其他表的hasone关系					
		},
        "many2manyTables": {
			// 中间表：staff_position，关联表：staff、position
            "staff_position": [		
                "staff",
                "position"
            ],
			...	// 其他表的many2many关系
        }
    }
}
```
详细配置文件参数说明：
```
	dbDriver		string				指定数据库引擎（mysql、postgres、sqlite、sqlserver），默认值：mysql
 	dbName  		string          	数据库名称  
	dsn     		string          	用于连接数据库的DSN  
	outPath 		string          	指定输出目录(默认值：./dao/query) 
	outFile 		string          	指定输出文件(默认值：gen.go)
	nspname 		string          	postgres数据库模式名称，默认值：public，如果数据库中的表不在public模式下，需要指定该参数
	dataMap 		map[string]string   	数据库自定义字段的数据类型
	fieldNullable 		bool   			表字段可为 null 值时, 对应结体字段使用指针类型，默认值：false
	fieldCoverable 		bool 			当字段具有默认值时生成指针，以解决无法分配零值的问题，默认值：false
	fieldSignable 		bool  			模型结构体字段的数字类型的符号表示是否与表字段的一致, false指示都用有符号类型，默认值：false
	fieldWithIndexTag 	bool  			生成 gorm 标签的字段索引属性，默认值：false
	fieldWithTypeTag 	bool 			生成 gorm 标签的字段类型属性，默认值：false
	withUnitTest 		bool  			生成单元测试，默认值：false,
	modelPkgPath 		string   		生成模型代码包名称。默认值：model  
	singularTable		bool			是否使用单数表名，默认值：true

	tables 				[]string 		指定要生成的表名，为空时生成数据库中所有表
	belongstoTables 	map[string][]string 	指定表的关联表，生成关联表的查询方法
	hasoneTables 		map[string][]string 	指定表的一对一关联表，生成关联表的查询方法
	many2manyTables 	map[string][]string 	指定表的多对多关联表，生成关联表的查询方法
```
```
	belongstoTables     key:表名（子表名），value:关联表名（主表名）
	hasoneTables        key:表名（主表名），value:关联表名（子表名）
	many2manyTables     key:表名（中间表名），value:关联表名（子表名，子表名）

```
```
注意事项：
    1、如果配置了tables数组，程序将只处理tables数组中的表及其关联关系，任何不在tables中的表，且跟tables中的表有关联关系的表都不会处理。
	2、如果没有配置tables数组，程序将处理数据库中的所有表及其关联关系。
	3、如果配置了tables数组，并且belongstoTables、hasoneTables、many2manyTables也有配置，那么belongstoTables、hasoneTables、many2manyTables中的所有表名必须包含在tables数组中，否则会报错。
	4、如果没有配置belongstoTables、hasoneTables、many2manyTables，那么数据库中所有设置了外键的表之间的关联关系默认为一对多（hasmany）关系。
```