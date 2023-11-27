package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var dbName = flag.String("dbName", "", "指定数据库名称")
var outPath = flag.String("outPath", "", "指定输出目录(默认 ./dao/query)")
var dsn = flag.String("dsn", "", "用于连接数据库的DSN  ")
var helpFlag = flag.Bool("h", false, "帮助文档")
var configFile = flag.String("c", "", "配置文件路径")

type Config struct {
	DbName  string            `json:"dbName"`
	Dsn     string            `json:"dsn"`
	OutPath string            `json:"outPath"`
	DataMap map[string]string `json:"dataMap"` // 自定义字段的数据类型

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
	// Genrated 查询代码文件名称，默认值：gen.go
	OutFile string `json:"outFile"`
	// 生成模型代码包名称。默认值：model
	ModelPkgPath string `json:"modelPkgPath"`
}

var configFromFile = Config{}

// readConfig 从文件中读取配置信息
func readConfig(filename string) (Config, error) {
	var config Config

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}

	// 将 JSON 解析到 Config 结构体
	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}
	fmt.Println("config:==================", config)

	return config, nil
}

// getValueOrDefault 返回非空值，如果为空，则返回默认值
func getValueOrDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 如果用户使用了 -h 参数，则显示帮助信息
	if *helpFlag {
		displayHelp()
		return
	}

	// 如果用户使用了 -c 参数，则读取配置文件
	// 读取配置文件（如果提供了配置文件选项）
	if *configFile != "" {
		config, err := readConfig(*configFile)
		if err != nil {
			fmt.Println("Error reading config:", err)
			return
		}
		configFromFile = config
	}

	// 如果用户使用了 -dsn 参数，则使用该参数值覆盖配置文件中的值
	*dsn = getValueOrDefault(*dsn, configFromFile.Dsn)

	// 使用命令行选项覆盖配置文件中的值
	*dbName = getValueOrDefault(*dbName, configFromFile.DbName)
	*outPath = getValueOrDefault(*outPath, configFromFile.OutPath)

	// 生成所有model和query
	processAllTables(initInfo())
	// 处理表关联关系
	processTableRelations(initInfo())
}

// 显示帮助信息的函数
func displayHelp() {
	fmt.Println("Usage: your_program [options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
	os.Exit(0)
}

/**
 * 初始化数据库连接
 * 生成generator实例
 * 自定义字段的数据类型
 * 自定义模型结体字段的标签
 */
func initInfo() (db *gorm.DB, g *gen.Generator, fieldOpts []gen.ModelOpt) {
	// var err error
	// 连接数据库
	db, err := gorm.Open(mysql.Open(*dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic(fmt.Errorf("数据库连接失败，请检查连接配置: %w", err))
	}

	if *outPath == "" {
		workDir, _ := os.Getwd()
		*outPath = workDir + "dao/query"
	}

	if configFromFile.ModelPkgPath == "" {
		configFromFile.ModelPkgPath = "model"
	}

	if configFromFile.OutFile == "" {
		configFromFile.OutFile = "gen.go"
	}
	// 生成实例
	g = gen.NewGenerator(gen.Config{
		// 相对执行`go run`时的路径, 会自动创建目录，相对路径为工程根目录
		OutPath: *outPath,

		// WithDefaultQuery 生成默认查询结构体(作为全局变量使用), 即`Q`结构体和其字段(各表模型)
		// WithoutContext 生成没有context调用限制的代码供查询
		// WithQueryInterface 生成interface形式的查询代码(可导出), 如`Where()`方法返回的就是一个可导出的接口类型
		Mode: gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,

		// 表字段可为 null 值时, 对应结体字段使用指针类型
		FieldNullable: configFromFile.FieldNullable,

		// 表字段默认值与模型结构体字段零值不一致的字段, 在插入数据时需要赋值该字段值为零值的, 结构体字段须是指针类型才能成功, 即`FieldCoverable:true`配置下生成的结构体字段.
		// 因为在插入时遇到字段为零值的会被GORM赋予默认值. 如字段`age`表默认值为10, 即使你显式设置为0最后也会被GORM设为10提交.
		// 如果该字段没有上面提到的插入时赋零值的特殊需要, 则字段为非指针类型使用起来会比较方便.
		FieldCoverable: configFromFile.FieldCoverable,

		// 模型结构体字段的数字类型的符号表示是否与表字段的一致, `false`指示都用有符号类型
		FieldSignable: configFromFile.FieldSignable,
		// 生成 gorm 标签的字段索引属性
		FieldWithIndexTag: configFromFile.FieldWithIndexTag,
		// 生成 gorm 标签的字段类型属性
		FieldWithTypeTag: configFromFile.FieldWithTypeTag,
		// 生成单元测试，默认值 false, 选项: false / true
		WithUnitTest: configFromFile.WithUnitTest,
		// 生成模型代码包名称。默认值：model
		ModelPkgPath: configFromFile.ModelPkgPath,
		// 生成的query code文件名称，默认值：gen.go
		OutFile: configFromFile.OutFile,
	})
	// 设置目标 db
	g.UseDB(db)

	// 自定义字段的数据类型
	// 统一数字类型为int64,兼容protobuf
	dataMap := map[string]func(columnType gorm.ColumnType) (dataType string){}
	if configFromFile.DataMap != nil {
		for k, v := range configFromFile.DataMap {
			dataMap[k] = func(columnType gorm.ColumnType) (dataType string) { return v }
		}
	}

	// 要先于`ApplyBasic`执行
	g.WithDataTypeMap(dataMap)

	// 自定义模型结体字段的标签
	// 将特定字段名的 json 标签加上`string`属性,即 MarshalJSON 时该字段由数字类型转成字符串类型
	// jsonField := gen.FieldJSONTagWithNS(func(columnName string) (tagContent string) {
	// 	toStringField := `balance, `
	// 	if strings.Contains(toStringField, columnName) {
	// 		return columnName + ",string"
	// 	}
	// 	return columnName
	// })

	// 将非默认字段名的字段定义为自动时间戳和软删除字段;
	// 自动时间戳默认字段名为:`updated_at`、`created_at, 表字段数据类型为: INT 或 DATETIME
	// 软删除默认字段名为:`deleted_at`, 表字段数据类型为: DATETIME
	autoUpdateTimeField := gen.FieldGORMTag("updatedAt", func(tag field.GormTag) field.GormTag {
		return tag.Append("autoUpdateTime")
	})
	autoCreateTimeField := gen.FieldGORMTag("createdAt", func(tag field.GormTag) field.GormTag {
		return tag.Append("autoCreateTime")
	})
	softDeleteField := gen.FieldType("deletedAt", "gorm.DeletedAt")

	// 模型自定义选项组
	fieldOpts = []gen.ModelOpt{
		// jsonField,
		autoCreateTimeField,
		autoUpdateTimeField,
		softDeleteField,
	}

	return db, g, fieldOpts
}

/**
 *创建全部模型文件，生成所有model和query
 *将生成的query目录下的gen.go文件移动到当前目录tmp文件夹下
 */
func processAllTables(db *gorm.DB, g *gen.Generator, fieldOpts []gen.ModelOpt) {
	allModel := g.GenerateAllTable(fieldOpts...)
	g.ApplyBasic(allModel...)
	g.Execute()

	// 将生成的query目录下的gen.go文件移动到当前目录tmp文件夹下
	moveGenFile()
}

type Results struct {
	TABLE_NAME             string //子表名
	COLUMN_NAME            string //子表列名
	CONSTRAINT_NAME        string //约束名
	REFERENCED_TABLE_NAME  string //关联表名
	REFERENCED_COLUMN_NAME string //关联列名
}

/**
 * 处理表关联关系
 */
func processTableRelations(db *gorm.DB, g *gen.Generator, fieldOpts []gen.ModelOpt) {
	relationList := []Results{}
	// 执行这条sql语句，获取当前数据库中所有表之间的外键关联关系
	// 执行结果保存到relationList中
	db.Raw("SELECT TABLE_NAME,COLUMN_NAME,CONSTRAINT_NAME,REFERENCED_TABLE_NAME,REFERENCED_COLUMN_NAME FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE WHERE TABLE_SCHEMA = ? AND REFERENCED_TABLE_SCHEMA IS NOT NULL;", *dbName).Scan(&relationList)

	type subTable struct {
		TABLE_NAME     string //子表名
		TABLE_NAME_UP  string //子表名首字母大写
		COLUMN_NAME    string //子表列名
		COLUMN_NAME_UP string //子表列名首字母大写
	}

	masterTableMap := make(map[string][]subTable)
	// 将relationList中的数据按照关联表名进行分组，将关联了父表名的所有子表数据放到一个切片中，然后将切片放到map中，map的key为父表名，value为子表切片
	for _, sub := range relationList {
		st := subTable{
			TABLE_NAME:     sub.TABLE_NAME,                         //子表名
			COLUMN_NAME:    sub.COLUMN_NAME,                        //子表列名
			TABLE_NAME_UP:  Case2Camel(sub.TABLE_NAME),             //将子表名下划线去掉，转换成首字母大写
			COLUMN_NAME_UP: Case2Camel(ProcessID(sub.COLUMN_NAME)), //将子表列名中以id结尾的字段中的id转换成ID格式，再将子表列名下划线去掉，转换成首字母大写
		}
		masterTableMap[sub.REFERENCED_TABLE_NAME] = append(masterTableMap[sub.REFERENCED_TABLE_NAME], st)
	}

	fmt.Println("主表 Map:::", ToJson(masterTableMap))

	// 生成新的generator实例，用于通过数据库子表名称，创建子表的模型基本结构体（BaseStruct）
	newGenerator := gen.NewGenerator(gen.Config{})
	newGenerator.UseDB(db)

	relationModels := []any{}
	// 遍历map，将map中的数据取出来，生成对应的关联关系模型文件
	for masterTable, subTables := range masterTableMap {
		subModels := []gen.ModelOpt{}
		// 遍历子表切片，将子表切片中的数据取出来，生成对应的关联关系模型文件
		for _, subTable := range subTables {
			// 目前只支持一对多关联关系，即：HasMany
			// 但是也能覆盖has_one和belongs_to的关联关系，只不过在生成的model中会多出一个切片字段，该切片中只有一个值
			// 对于多对多关联关系(many2many)，请先设计中间连接表，连接表中定义两个主键，即：复合主键，每个主键关联一张主表，
			// 这样就能生成两个一对多的关联关系，再运行本程序，就能实现多对多的关联关系了
			subModels = append(subModels, gen.FieldRelate(field.HasMany, subTable.TABLE_NAME_UP, newGenerator.GenerateModel(subTable.TABLE_NAME),
				&field.RelateConfig{
					// RelateSlice配置为true，那么在主表生成model的时候会生成关联表的切片
					RelateSlice: true,
					// 配置关联关系的外键字段，并且将外键字段的gorm标签中的foreignKey属性设置为关联表的列名
					GORMTag: field.GormTag{"foreignKey": {subTable.COLUMN_NAME_UP}},
				}))
		}
		relationModels = append(relationModels, g.GenerateModel(masterTable, append(fieldOpts, subModels...)...))
	}

	g.ApplyBasic(relationModels...)
	g.Execute()

	// 将当前目录tmp文件夹下的gen.go文件移动到query目录下
	moveGenFileBack()
}

/**
 * 将生成的query目录下的gen.go根文件移动到当前目录tmp文件夹下，
 * gen.go文件中保存的是所有表的模型的引用，
 * gen在生成query文件时，只会将ApplyBasic方法参数中的模型写入query中的根文件gen.go中，
 * 而我们在后续调用processTableRelations方法处理关联关系的时候，只处理有关联关系的表，
 * 方法中生成的gen.go中只会有有关联关系的表的模型的引用，因此需要将保存了所有表的模型的引用的gen.go文件
 * 移动到tmp文件夹下，然后再调用processTableRelations方法处理关联关系，处理完关联关系后，
 * 再将tmp文件夹下的gen.go文件移动到query目录下。
 */
func moveGenFile() {
	workDir, _ := os.Getwd()
	err := os.MkdirAll(workDir+"/gen_tool_plus_tmp", 0777)
	if err != nil {
		fmt.Println("创建文件夹logs失败!", err)
		return
	}
	genFile := *outPath + "/gen.go"
	if _, err := os.Stat(genFile); err != nil {
		fmt.Println("moveGenFile:", genFile)
		fmt.Println("gen.go文件不存在!")
		return
	}
	fmt.Println("gen.go文件存在:", genFile)
	os.Rename(genFile, workDir+"/gen_tool_plus_tmp/gen.go")
}

/**
 * 将当前目录tmp文件夹下的gen.go文件移动到query目录下
 */
func moveGenFileBack() {
	workDir, _ := os.Getwd()
	genFile := *outPath + "/gen.go"

	// 删除临时创建的gen_tool_plus_tmp文件夹
	defer deleteTmpDir()

	if _, err := os.Stat(genFile); err != nil {
		fmt.Println("moveGenFileBack:", genFile)
		fmt.Println("gen.go文件不存在!")
		return
	}
	err := os.Rename(workDir+"/gen_tool_plus_tmp/gen.go", genFile)
	if err != nil {
		fmt.Println("移动文件失败!", err)
		return
	}
}

/**
*删除临时创建的tmp文件夹
 */
func deleteTmpDir() {
	workDir, _ := os.Getwd()
	// 要删除的文件夹路径
	folderPath := workDir + "/gen_tool_plus_tmp"

	// 删除文件夹
	err := os.RemoveAll(folderPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("临时gen_tool_plus_tmp文件夹删除成功")
}

// 下划线写法转为驼峰写法
func Case2Camel(name string) string {
	words := strings.Split(name, "_")
	var result string
	for _, word := range words {
		result += strings.ToUpper(string(word[0])) + word[1:]
	}
	return result
}

func ProcessID(str string) string {
	if strings.HasSuffix(str, "id") {
		str, _ = strings.CutSuffix(str, "id")
		str = str + "ID"
	}
	return str
}

func ToJson(result interface{}) string {
	jsonBytes, _ := json.Marshal(result)
	return string(jsonBytes)
}
