package common

import "gorm.io/gen/field"

type ConfigFile struct {
	Version  string   `json:"version"`
	Database DBConfig `json:"database"`
}

type DBConfig struct {
	DbDriver string            `json:"dbDriver"` // 数据库驱动
	DbName   string            `json:"dbName"`
	Dsn      string            `json:"dsn"`
	OutPath  string            `json:"outPath"`
	OutFile  string            `json:"outFile"`
	DataMap  map[string]string `json:"dataMap"` // 自定义字段的数据类型

	Tables []string `json:"tables"` // 指定要生成的表名

	// 表字段可为 null 值时, 对应结体字段使用指针类型
	FieldNullable bool `json:"fieldNullable"`

	// 表字段默认值与模型结构体字段零值不一致的字段, 在插入数据时需要赋值该字段值为零值的, 结构体字段须是指针类型才能成功, 即`FieldCoverable:true`配置下生成的结构体字段.
	// 因为在插入时遇到字段为零值的会被GORM赋予默认值. 如字段`age`表默认值为10, 即使你显式设置为0最后也会被GORM设为10提交.
	// 如果该字段没有上面提到的插入时赋零值的特殊需要, 则字段为非指针类型使用起来会比较方便.
	FieldCoverable bool `json:"fieldCoverable"`

	// 模型结构体字段的数字类型的符号表示是否与表字段的一致, `false`指示都用有符号类型
	FieldSignable bool `json:"fieldSignable"`
	// 生成 gorm 标签的字段索引属性
	FieldWithIndexTag bool `json:"fieldWithIndexTag"`
	// 生成 gorm 标签的字段类型属性
	FieldWithTypeTag bool `json:"fieldWithTypeTag"`
	// 生成单元测试，默认值 false, 选项: false / true
	WithUnitTest bool `json:"withUnitTest"`
	// 生成模型代码包名称。默认值：model
	ModelPkgPath string `json:"modelPkgPath"`
	// 表名单数形式，即表名不加s后缀。默认值 true, 选项: false / true
	SingularTable bool `json:"singularTable"`
	// belongsto关联关系
	BelongstoTables map[string][]string `json:"belongstoTables"`
	// hasone关联关系
	HasoneTables map[string][]string `json:"hasoneTables"`
	// many2many关联关系
	Many2manyTables map[string][]string `json:"many2manyTables"`
	// postgres数据库中的schema名称
	Nspname string `json:"nspname"`
	//	json tag 命名格式 默认为false，即与数据库表字段保持一致，true为使用驼峰命名
	JsonTagFormat bool `json:"jsonTagFormat"`
}

// Results 存储数据库关联关系查询结果
type Results struct {
	TABLE_NAME             string //子表名
	COLUMN_NAME            string //子表列名
	REFERENCED_TABLE_NAME  string //关联表名
	REFERENCED_COLUMN_NAME string //关联列名
}

// SubTable 格式化后的数据库关联关系查询结果及关联关系类型
type SubTable struct {
	TABLE_NAME               string                 //子表名
	TABLE_NAME_UP            string                 //子表名首字母大写
	COLUMN_NAME              string                 //子表列名
	COLUMN_NAME_UP           string                 //子表列名首字母大写
	REFERENCED_TABLE_NAME    string                 //关联表名
	REFERENCED_TABLE_NAME_UP string                 //关联表名首字母大写
	RELATION_TYPE            field.RelationshipType //关联关系类型
	MIDDLE_TABLE             string                 //中间表名
}
