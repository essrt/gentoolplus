package global

import (
	"github.com/essrt/gentoolplus/common"
	"gorm.io/gorm"
)

var (
	DB       *gorm.DB                                  // 数据库连接
	Config   *common.ConfigFile = &common.ConfigFile{} // 配置文件
	DbName   *string                                   // 数据库名称
	Dsn      *string                                   // 数据库连接字符串
	OutPath  *string                                   // 输出目录
	OutFile  *string                                   // 输出文件
	DbDriver *string                                   // 数据库驱动
)
