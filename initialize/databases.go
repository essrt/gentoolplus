package initialize

import (
	"fmt"
	"strings"
	"sync"

	"github.com/essrt/gentoolplus/global"
	"github.com/essrt/gentoolplus/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	once sync.Once
)

func init() {

	// 初始化配置文件
	initConfig()
	// 初始化数据库连接
	initDB()
	// 检查配置文件中的表名在数据库中是否存在
	checkDbTables()
}

func initDB() {
	once.Do(func() {
		var err error
		var dial gorm.Dialector = mysql.Open(*global.Dsn)

		if *global.DbDriver == "mysql" {
			dial = mysql.Open(*global.Dsn)
		} else if *global.DbDriver == "postgres" {
			dial = postgres.Open(*global.Dsn)
		} else if *global.DbDriver == "sqlite" {
			dial = sqlite.Open(*global.Dsn)
		} else if *global.DbDriver == "sqlserver" {
			dial = sqlserver.Open(*global.Dsn)
		} else {
			panic(fmt.Errorf("不支持的数据库类型: %w", err))
		}

		global.DB, err = gorm.Open(dial, &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: global.Config.Database.SingularTable,
			},
		})
		if err != nil {
			panic(fmt.Errorf("数据库连接失败，请检查连接配置: %w", err))
		}
	})
}

// checkDbTables 检查配置文件中的表名在数据库中是否存在
func checkDbTables() {

	// 配置文件中hasone、belongsto、many2many关系表名称的切片
	configTables := []string{}
	tableNames := []string{}
	if *global.DbDriver == "mysql" {
		global.DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = ?;", *global.DbName).Scan(&tableNames)
	} else if *global.DbDriver == "postgres" {
		global.DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_catalog = ?;", *global.DbName).Scan(&tableNames)
	} else if *global.DbDriver == "sqlite" {
		global.DB.Raw("SELECT name AS table_name FROM sqlite_master WHERE type = 'table';").Scan(&tableNames)
	} else if *global.DbDriver == "sqlserver" {
		global.DB.Raw("USE " + *global.DbName + "; SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE';").Scan(&tableNames)
	} else {
		panic(fmt.Errorf("不支持的数据库类型: %s", *global.DbDriver))
	}

	tmp := []string{}
	// 去掉hasone关系表名称中的字符串中的空格或者换行符
	if global.Config.Database.HasoneTables != nil {
		newHasoneTables := make(map[string][]string)
		for key, values := range global.Config.Database.HasoneTables {
			key = strings.TrimSpace(key)
			configTables = append(configTables, key)
			tmp = append(tmp, key)
			for _, value := range values {
				value = strings.TrimSpace(value)
				newHasoneTables[key] = append(newHasoneTables[key], value)
				configTables = append(configTables, value) // 将配置文件中的表名放到切片中
				tmp = append(tmp, value)
			}
		}
		global.Config.Database.HasoneTables = newHasoneTables
	}

	// 去掉belongsto关系表名称中的字符串中的空格或者换行符
	if global.Config.Database.BelongstoTables != nil {
		newBelongstoTables := make(map[string][]string)
		for key, values := range global.Config.Database.BelongstoTables {
			key = strings.TrimSpace(key)
			configTables = append(configTables, key)
			tmp = append(tmp, key)
			for _, value := range values {
				value = strings.TrimSpace(value)
				newBelongstoTables[key] = append(newBelongstoTables[key], value)
				configTables = append(configTables, value) // 将配置文件中的表名放到切片中
				tmp = append(tmp, value)
			}
		}
		global.Config.Database.BelongstoTables = newBelongstoTables
	}

	// 去掉many2many关系表名称中的字符串中的空格或者换行符
	if global.Config.Database.Many2manyTables != nil {
		newMany2manyTables := make(map[string][]string)
		for key, values := range global.Config.Database.Many2manyTables {
			// 检查配置文件中的many2manyTables配置是否正确，many2manyTables配置的关联关系必须是2个表之间的关联关系
			if len(values) != 2 {
				err := fmt.Errorf("配置文件错误：many2manyTables配置错误，配置项的value值必须是关联的2个表！")
				panic(err)
			}
			key = strings.TrimSpace(key)
			configTables = append(configTables, key)
			tmp = append(tmp, key)
			for _, value := range values {
				value = strings.TrimSpace(value)
				newMany2manyTables[key] = append(newMany2manyTables[key], value)
				configTables = append(configTables, value) // 将配置文件中的表名放到切片中
				tmp = append(tmp, value)
			}
		}
		global.Config.Database.Many2manyTables = newMany2manyTables
	}

	if len(tmp) > 0 && global.Config.Database.Tables != nil && len(global.Config.Database.Tables) > 0 {
		for _, table := range tmp {
			if !utils.ContainsValue(global.Config.Database.Tables, table) {
				err := fmt.Errorf("配置文件错误：表名 %s 不在tables配置项中！", table)
				panic(err)
			}
		}
	}

	if global.Config.Database.Tables != nil && len(global.Config.Database.Tables) > 0 {
		configTables = append(configTables, global.Config.Database.Tables...)
	}

	// 检查配置文件中的表名是否存在在数据库中
	for _, configTable := range configTables {
		if !utils.ContainsValue(tableNames, configTable) {
			err := fmt.Errorf("配置文件错误：表名 %s 不在数据库中！", configTable)
			panic(err)
		}
	}
}
