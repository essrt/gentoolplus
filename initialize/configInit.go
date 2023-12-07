package initialize

import (
	"flag"
	"fmt"
	"os"

	"github.com/essrt/gentoolplus/global"
	"github.com/essrt/gentoolplus/utils"
	"github.com/spf13/viper"
)

var helpFlag = flag.Bool("h", false, "帮助文档")
var configFile = flag.String("c", "", "配置文件路径")

func initConfig() {
	global.DbName = flag.String("dbName", "", "指定数据库名称")
	global.OutPath = flag.String("outPath", "", "指定输出目录(默认 ./dao/query)")
	global.OutFile = flag.String("outFile", "", "指定输出文件(默认 gen.go)")
	global.Dsn = flag.String("dsn", "", "用于连接数据库的DSN  ")
	global.DbDriver = flag.String("dbDriver", "", "数据库驱动")

	flag.Parse()

	// 如果用户使用了 -h 参数，则显示帮助信息
	if *helpFlag {
		displayHelp()
		return
	}

	// 如果用户使用了 -c 参数，则读取配置文件
	// 读取配置文件（如果提供了配置文件选项）
	if *configFile != "" {
		err := readConfig(*configFile)
		if err != nil {
			fmt.Println("读取配置文件失败: %w", err)
			return
		}

		fmt.Println("配置文件信息:", utils.ToJson(global.Config))
	}

	// 如果用户使用了 -dsn 参数，则使用该参数值覆盖配置文件中的值
	*global.Dsn = getValueOrDefault(*global.Dsn, global.Config.Database.Dsn)

	// 使用命令行选项覆盖配置文件中的值
	*global.DbName = getValueOrDefault(*global.DbName, global.Config.Database.DbName)
	*global.OutPath = getValueOrDefault(*global.OutPath, global.Config.Database.OutPath)
	*global.OutFile = getValueOrDefault(*global.OutFile, global.Config.Database.OutFile)
	*global.DbDriver = getValueOrDefault(*global.DbDriver, global.Config.Database.DbDriver)
}

// 显示帮助信息的函数
func displayHelp() {
	fmt.Println("用法：gentoolplus [选项]")
	fmt.Println("选项：")
	flag.PrintDefaults()
	os.Exit(0)
}

// readConfig 从文件中读取配置信息
func readConfig(filename string) error {
	v := viper.New()

	v.SetDefault("database.dbDriver", "mysql")
	v.SetDefault("database.outPath", "./dao/query")
	v.SetDefault("database.outFile", "gen.go")
	v.SetDefault("database.fieldNullable", true)
	v.SetDefault("database.fieldCoverable", true)
	v.SetDefault("database.fieldSignable", false)
	v.SetDefault("database.fieldWithIndexTag", false)
	v.SetDefault("database.fieldWithTypeTag", false)
	v.SetDefault("database.withUnitTest", false)
	v.SetDefault("database.singularTable", true)
	v.SetDefault("database.nspname", "public")
	v.SetDefault("database.modelPkgPath", "model")

	// 设置配置文件的名称和类型
	v.SetConfigName("gentoolplus_config")
	v.SetConfigType("json")

	//文件的路径设置
	v.SetConfigFile(filename)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	err := v.Unmarshal(global.Config)

	if err != nil {
		fmt.Println("读取配置失败")
		return err
	}
	return nil
}

// getValueOrDefault 返回非空值，如果为空，则返回默认值
func getValueOrDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}
