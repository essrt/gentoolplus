package process

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/essrt/gentoolplus/common"
	"github.com/essrt/gentoolplus/global"
	"github.com/essrt/gentoolplus/utils"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

/**
 *创建全部模型文件，生成所有model和query
 *将生成的query目录下的gen.go文件移动到当前目录tmp文件夹下
 */
func ProcessAllTables() {
	g, fieldOpts := utils.InitGenGenerator()
	allModel := []any{}
	config := global.Config.Database
	if config.Tables != nil && len(config.Tables) > 0 {
		for _, table := range config.Tables {
			allModel = append(allModel, g.GenerateModel(table, fieldOpts...))
		}
	} else {
		allModel = g.GenerateAllTable(fieldOpts...)
	}

	g.ApplyBasic(allModel...)
	g.Execute()

	// 将生成的query目录下的gen.go文件移动到当前目录tmp文件夹下
	utils.MoveGenFile()
}

// 处理sqlite数据库中的表关联关系
func ProcessSqliteRelation() (relationList []common.Results) {
	// 打开 SQLite 数据库连接
	db, err := sql.Open("sqlite3", *global.Dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// 查询所有表
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table';")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// 遍历每个表
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatal(err)
		}

		// 查询表的外键关系
		fkRows, err := db.Query(fmt.Sprintf("PRAGMA foreign_key_list(%s);", tableName))
		if err != nil {
			log.Fatal(err)
		}
		defer fkRows.Close()

		// 遍历外键关系
		for fkRows.Next() {
			var id, seq, table, from, to, on_update, on_delete, match string
			if err := fkRows.Scan(&id, &seq, &table, &from, &to, &on_update, &on_delete, &match); err != nil {
				log.Fatal(err)
			}
			relationList = append(relationList, common.Results{TABLE_NAME: tableName, COLUMN_NAME: from, REFERENCED_TABLE_NAME: table, REFERENCED_COLUMN_NAME: to})
		}
	}

	return relationList
}

/**
 * 处理表关联关系
 */
func ProcessTableRelations() {
	g, fieldOpts := utils.InitGenGenerator()
	relationList := []common.Results{}
	config := global.Config.Database
	// 执行这条sql语句，获取当前数据库中所有表之间的外键关联关系
	// 执行结果保存到relationList中
	if *global.DbDriver == "mysql" {
		global.DB.Raw("SELECT TABLE_NAME,COLUMN_NAME,CONSTRAINT_NAME,REFERENCED_TABLE_NAME,REFERENCED_COLUMN_NAME FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE WHERE TABLE_SCHEMA = ? AND REFERENCED_TABLE_SCHEMA IS NOT NULL;", *global.DbName).Scan(&relationList)
	} else if *global.DbDriver == "postgres" {
		global.DB.Raw("SELECT conname AS constraint_name, conrelid::regclass AS table_name, a.attname AS column_name, confrelid::regclass AS referenced_table_name, af.attname AS referenced_column_name FROM pg_constraint c JOIN pg_attribute a ON a.attnum = ANY(c.conkey) AND a.attrelid = c.conrelid JOIN pg_attribute af ON af.attnum = ANY(c.confkey) AND af.attrelid = c.confrelid WHERE c.confrelid IS NOT NULL AND c.connamespace = (SELECT oid FROM pg_namespace WHERE nspname = ?);", config.Nspname).Scan(&relationList)
	} else if *global.DbDriver == "sqlite" {
		relationList = ProcessSqliteRelation()
	} else if *global.DbDriver == "sqlserver" {
		global.DB.Raw("USE " + *global.DbName + "; SELECT t.name AS TABLE_NAME, col.name AS COLUMN_NAME, fk.name AS ForeignKeyName, ref.name AS REFERENCED_TABLE_NAME, refCol.name AS REFERENCED_COLUMN_NAME FROM sys.tables AS t INNER JOIN sys.foreign_keys AS fk ON t.object_id = fk.parent_object_id INNER JOIN sys.foreign_key_columns AS fkc ON fk.object_id = fkc.constraint_object_id INNER JOIN sys.columns AS col ON fkc.parent_column_id = col.column_id AND fkc.parent_object_id = col.object_id INNER JOIN sys.tables AS ref ON fk.referenced_object_id = ref.object_id INNER JOIN sys.columns AS refCol ON fkc.referenced_column_id = refCol.column_id AND fkc.referenced_object_id = refCol.object_id;").Scan(&relationList)
	} else {
		panic(fmt.Errorf("不支持的数据库类型: %s", *global.DbDriver))
	}

	// hasOne关系列表
	var hasOneRelationList []string
	// belongsTo关系列表
	var belongsToRelationList []string

	if config.HasoneTables != nil {
		for key, values := range config.HasoneTables {
			for _, value := range values {
				hasOneRelationList = append(hasOneRelationList, key+"_"+value)
			}
		}
	}

	if config.BelongstoTables != nil {
		for key, values := range config.BelongstoTables {
			for _, value := range values {
				belongsToRelationList = append(belongsToRelationList, value+"_"+key)
			}
		}
	}

	tmpRelationList := []string{}
	for _, v := range relationList {
		if config.Tables != nil && len(config.Tables) > 0 {
			if utils.ContainsValue(config.Tables, v.TABLE_NAME) && utils.ContainsValue(config.Tables, v.REFERENCED_TABLE_NAME) {
				tmpRelationList = append(tmpRelationList, v.REFERENCED_TABLE_NAME+"_"+v.TABLE_NAME)
			}
		} else {
			tmpRelationList = append(tmpRelationList, v.REFERENCED_TABLE_NAME+"_"+v.TABLE_NAME)
		}
	}

	finalRelationList := []common.Results{}
	// 如果配置文件中指定了要生成的表名，则只生成指定的表名之间的关联关系
	if config.Tables != nil && len(config.Tables) > 0 {
		for _, v := range relationList {
			if utils.ContainsValue(config.Tables, v.TABLE_NAME) && utils.ContainsValue(config.Tables, v.REFERENCED_TABLE_NAME) {
				finalRelationList = append(finalRelationList, v)
			}
		}
	} else {
		finalRelationList = relationList
	}

	// 检查hasonerelationList和belongstoRelationList中的值在relationList中是否存在
	for _, v := range hasOneRelationList {
		if !utils.ContainsValue(tmpRelationList, v) {
			err := fmt.Errorf("配置文件错误：配置项 hasOneRelationList 中，%s不存在关联关系！", v)
			panic(err)
		}
	}

	for _, v := range belongsToRelationList {
		if !utils.ContainsValue(tmpRelationList, v) {
			err := fmt.Errorf("配置文件错误：配置项 belongsToRelationList 中，%s不存在关联关系！", v)
			panic(err)
		}
	}

	masterTableMap := make(map[string][]common.SubTable)
	// 将finalRelationList中的数据按照关联表名进行分组，将关联了父表名的所有子表数据放到一个切片中，然后将切片放到map中，map的key为父表名，value为子表切片
	for _, sub := range finalRelationList {

		st := common.SubTable{
			TABLE_NAME:               sub.TABLE_NAME,                                     //子表名
			COLUMN_NAME:              sub.COLUMN_NAME,                                    //子表列名
			TABLE_NAME_UP:            utils.Case2Camel(sub.TABLE_NAME),                   //将子表名下划线去掉，转换成首字母大写
			COLUMN_NAME_UP:           utils.Case2Camel(utils.ProcessID(sub.COLUMN_NAME)), //将子表列名中以id结尾的字段中的id转换成ID格式，再将子表列名下划线去掉，转换成首字母大写
			REFERENCED_TABLE_NAME:    sub.REFERENCED_TABLE_NAME,                          //关联表名
			REFERENCED_TABLE_NAME_UP: utils.Case2Camel(sub.REFERENCED_TABLE_NAME),        //将关联表名下划线去掉，转换成首字母大写
			RELATION_TYPE:            field.HasMany,                                      //关联关系类型
		}

		if utils.ContainsValue(hasOneRelationList, sub.REFERENCED_TABLE_NAME+"_"+sub.TABLE_NAME) {
			st.RELATION_TYPE = field.HasOne
			masterTableMap[sub.REFERENCED_TABLE_NAME] = append(masterTableMap[sub.REFERENCED_TABLE_NAME], st)
		} else if utils.ContainsValue(belongsToRelationList, sub.REFERENCED_TABLE_NAME+"_"+sub.TABLE_NAME) {
			st1 := common.SubTable{
				TABLE_NAME:               sub.REFERENCED_TABLE_NAME,                                     //子表名
				COLUMN_NAME:              sub.REFERENCED_COLUMN_NAME,                                    //子表列名
				TABLE_NAME_UP:            utils.Case2Camel(sub.REFERENCED_TABLE_NAME),                   //将子表名下划线去掉，转换成首字母大写
				COLUMN_NAME_UP:           utils.Case2Camel(utils.ProcessID(sub.REFERENCED_COLUMN_NAME)), //将子表列名中以id结尾的字段中的id转换成ID格式，再将子表列名下划线去掉，转换成首字母大写
				REFERENCED_TABLE_NAME:    sub.TABLE_NAME,                                                //关联表名
				REFERENCED_TABLE_NAME_UP: utils.Case2Camel(sub.TABLE_NAME),                              //将关联表名下划线去掉，转换成首字母大写
				RELATION_TYPE:            field.BelongsTo,                                               //关联关系类型
			}
			masterTableMap[sub.TABLE_NAME] = append(masterTableMap[sub.TABLE_NAME], st1)
		} else {
			masterTableMap[sub.REFERENCED_TABLE_NAME] = append(masterTableMap[sub.REFERENCED_TABLE_NAME], st)
		}
	}

	if config.Many2manyTables != nil {
		for middleTable, v := range config.Many2manyTables {

			st2 := common.SubTable{
				TABLE_NAME:               v[1],                   //子表名
				TABLE_NAME_UP:            utils.Case2Camel(v[1]), //将子表名下划线去掉，转换成首字母大写
				REFERENCED_TABLE_NAME:    v[0],                   //关联表名
				REFERENCED_TABLE_NAME_UP: utils.Case2Camel(v[0]),
				RELATION_TYPE:            field.Many2Many, //关联关系类型
				MIDDLE_TABLE:             middleTable,     //中间表名
			}

			st3 := common.SubTable{
				TABLE_NAME:               v[0],                   //子表名
				TABLE_NAME_UP:            utils.Case2Camel(v[0]), //将子表名下划线去掉，转换成首字母大写
				REFERENCED_TABLE_NAME:    v[1],                   //关联表名
				REFERENCED_TABLE_NAME_UP: utils.Case2Camel(v[1]),
				RELATION_TYPE:            field.Many2Many, //关联关系类型
				MIDDLE_TABLE:             middleTable,     //中间表名
			}

			masterTableMap[v[0]] = append(masterTableMap[v[0]], st2)
			masterTableMap[v[1]] = append(masterTableMap[v[1]], st3)
		}
	}

	tmp := make(map[string][]string)
	// 检查表之间是否存在循环关联关系
	for masterTable, subTables := range masterTableMap {
		tmpSubTables := []string{}
		for _, subTable := range subTables {
			tmpSubTables = append(tmpSubTables, subTable.TABLE_NAME)
		}
		tmp[masterTable] = tmpSubTables
		if table, exits := utils.HasDuplicate(tmpSubTables); exits {
			panic(fmt.Errorf("配置文件或数据库配置错误：表 %s 与表 %s 存在循环关联关系！", masterTable, table))
		}
	}

	fmt.Println("=========主表 Map:::", utils.ToJson(tmp))

	// 生成新的generator实例，用于通过数据库子表名称，创建子表的模型基本结构体（BaseStruct）
	newGenerator := gen.NewGenerator(gen.Config{})
	newGenerator.UseDB(global.DB)

	fmt.Println("主表 Map:::", utils.ToJson(masterTableMap))

	relationModels := []any{}
	// 遍历map，将map中的数据取出来，生成对应的关联关系模型文件
	for masterTable, subTables := range masterTableMap {
		subModels := []gen.ModelOpt{}
		// 遍历子表切片，将子表切片中的数据取出来，生成对应的关联关系模型文件
		for _, subTable := range subTables {
			if subTable.RELATION_TYPE == field.Many2Many {
				subModels = append(subModels, gen.FieldRelate(subTable.RELATION_TYPE, subTable.TABLE_NAME_UP, newGenerator.GenerateModel(subTable.TABLE_NAME),
					&field.RelateConfig{
						//
						GORMTag: field.GormTag{"many2many": {subTable.MIDDLE_TABLE}},
					}))
			} else {
				subModels = append(subModels, gen.FieldRelate(subTable.RELATION_TYPE, subTable.TABLE_NAME_UP, newGenerator.GenerateModel(subTable.TABLE_NAME),
					&field.RelateConfig{
						// 配置关联关系的外键字段，并且将外键字段的gorm标签中的foreignKey属性设置为关联表的列名
						GORMTag: field.GormTag{"foreignKey": {subTable.COLUMN_NAME_UP}},
					}))
			}
		}
		relationModels = append(relationModels, g.GenerateModel(masterTable, append(fieldOpts, subModels...)...))
	}

	g.ApplyBasic(relationModels...)
	g.Execute()

	// 将当前目录tmp文件夹下的gen.go文件移动到query目录下
	utils.MoveGenFileBack()
}
