package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/essrt/gentoolplus/global"
)

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

// 判断切片中是否存在某个值
func ContainsValue(slice []string, value string) bool {
	for _, element := range slice {
		if element == value {
			return true
		}
	}
	return false
}

// 判断切片中是否存在重复的值，并且返回重复的值
func HasDuplicate(slice []string) (string, bool) {
	for i := 0; i < len(slice); i++ {
		for j := i + 1; j < len(slice); j++ {
			if slice[i] == slice[j] {
				return slice[i], true
			}
		}
	}
	return "", false
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
func MoveGenFile() {
	workDir, _ := os.Getwd()
	err := os.MkdirAll(path.Join(workDir, "gen_tool_plus_tmp"), 0777)
	if err != nil {
		fmt.Println("创建文件夹logs失败!", err)
		return
	}
	genFile := path.Join(*global.OutPath, *global.OutFile)
	if _, err := os.Stat(genFile); err != nil {
		fmt.Println("moveGenFile:", genFile)
		fmt.Println("gen.go文件不存在!", err)
		return
	}
	fmt.Println("gen.go文件存在:", genFile)
	os.Rename(genFile, path.Join(workDir, "gen_tool_plus_tmp", *global.OutFile))
}

/**
 * 将当前目录tmp文件夹下的gen.go文件移动到query目录下
 */
func MoveGenFileBack() {
	workDir, _ := os.Getwd()
	genFile := path.Join(*global.OutPath, *global.OutFile)

	// 删除临时创建的gen_tool_plus_tmp文件夹
	defer DeleteTmpDir()

	if _, err := os.Stat(genFile); err != nil {
		fmt.Println("moveGenFileBack:", genFile)
		fmt.Println("gen.go文件不存在!", err)
		return
	}
	err := os.Rename(path.Join(workDir, "gen_tool_plus_tmp", *global.OutFile), genFile)

	if err != nil {
		fmt.Println("移动文件失败!", err)
		return
	}
}

/**
*删除临时创建的tmp文件夹
 */
func DeleteTmpDir() {
	workDir, _ := os.Getwd()
	// 要删除的文件夹路径
	folderPath := workDir + "/gen_tool_plus_tmp"

	// 删除文件夹
	err := os.RemoveAll(folderPath)
	if err != nil {
		fmt.Println("删除文件夹出错:", err)
		return
	}

	fmt.Println("临时gen_tool_plus_tmp文件夹删除成功")
}
