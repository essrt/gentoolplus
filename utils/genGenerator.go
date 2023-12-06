package utils

import (
	"github.com/essrt/gentoolplus/global"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

func InitGenGenerator() (g *gen.Generator, fieldOpts []gen.ModelOpt) {
	// 生成实例
	g = gen.NewGenerator(gen.Config{
		// 相对执行`go run`时的路径, 会自动创建目录，相对路径为工程根目录
		OutPath: *global.OutPath,
		OutFile: *global.OutFile,
		// WithDefaultQuery 生成默认查询结构体(作为全局变量使用), 即`Q`结构体和其字段(各表模型)
		// WithoutContext 生成没有context调用限制的代码供查询
		// WithQueryInterface 生成interface形式的查询代码(可导出), 如`Where()`方法返回的就是一个可导出的接口类型
		Mode: gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,

		// 表字段可为 null 值时, 对应结体字段使用指针类型
		FieldNullable: global.Config.Database.FieldNullable,

		// 表字段默认值与模型结构体字段零值不一致的字段, 在插入数据时需要赋值该字段值为零值的, 结构体字段须是指针类型才能成功, 即`FieldCoverable:true`配置下生成的结构体字段.
		// 因为在插入时遇到字段为零值的会被GORM赋予默认值. 如字段`age`表默认值为10, 即使你显式设置为0最后也会被GORM设为10提交.
		// 如果该字段没有上面提到的插入时赋零值的特殊需要, 则字段为非指针类型使用起来会比较方便.
		FieldCoverable: global.Config.Database.FieldCoverable,

		// 模型结构体字段的数字类型的符号表示是否与表字段的一致, `false`指示都用有符号类型
		FieldSignable: global.Config.Database.FieldSignable,
		// 生成 gorm 标签的字段索引属性
		FieldWithIndexTag: global.Config.Database.FieldWithIndexTag,
		// 生成 gorm 标签的字段类型属性
		FieldWithTypeTag: global.Config.Database.FieldWithTypeTag,
		// 生成单元测试，默认值 false, 选项: false / true
		WithUnitTest: global.Config.Database.WithUnitTest,
		// 生成模型代码包名称。默认值：model
		ModelPkgPath: global.Config.Database.ModelPkgPath,
	})
	// 设置目标 db
	g.UseDB(global.DB)

	// 自定义字段的数据类型
	// 统一数字类型为int64,兼容protobuf
	dataMap := map[string]func(columnType gorm.ColumnType) (dataType string){}
	if global.Config.Database.DataMap != nil {
		for k, v := range global.Config.Database.DataMap {
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

	return g, fieldOpts
}
