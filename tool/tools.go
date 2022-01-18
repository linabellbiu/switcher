package tool

import (
	"regexp"
	"sort"
	"strings"
)

func In(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	if index < len(strArray) && strArray[index] == target {
		return true
	}
	return false
}

func DelSpace(str string) string {
	// 多个空格替删除
	re, _ := regexp.Compile("\\s+")
	str = re.ReplaceAllString(str, "")
	return str
}

func DelExtraSpace(str string) string {
	// 多个空格替删除
	re, _ := regexp.Compile("\\s{2,}")
	str = strings.Trim(re.ReplaceAllString(str, " "), " ")
	return str
}
