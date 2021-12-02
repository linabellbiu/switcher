package parse

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

func delSpace(str string) string {
	//多个空格替删除
	re, _ := regexp.Compile("\\s+")
	str = re.ReplaceAllString(str, "")
	return str
}

func delExtraSpace(str string) string {
	//多个空格替删除
	re, _ := regexp.Compile("\\s{2,}")
	str = strings.Trim(re.ReplaceAllString(str, " "), " ")
	return str
}

/*
	转换为大驼峰命名法则
	首字母大写，“_” 忽略后大写
*/
func marshal(name string) string {
	if name == "" {
		return ""
	}
	temp := strings.Split(name, "_")
	var s string
	for _, v := range temp {
		vv := []rune(v)
		if len(vv) > 0 {
			if bool(vv[0] >= 'a' && vv[0] <= 'z') { //首字母大写
				vv[0] -= 32
			}
			s += string(vv)
		}
	}
	s = uncommonInitialismsReplacer.Replace(s)
	return s
}

// Copied from golint
var commonInitialisms []string

//var commonInitialisms = []string{"ACL", "API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SQL", "SSH", "TCP", "TLS", "TTL", "UDP", "UI", "UID", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XMPP", "XSRF", "XSS"}
var uncommonInitialismsReplacer *strings.Replacer

func init() {
	var commonInitialismsForReplacer []string
	var uncommonInitialismsForReplacer []string
	for _, initialism := range commonInitialisms {
		commonInitialismsForReplacer = append(commonInitialismsForReplacer, initialism, strings.Title(strings.ToLower(initialism)))
		uncommonInitialismsForReplacer = append(uncommonInitialismsForReplacer, strings.Title(strings.ToLower(initialism)), initialism)
	}
	uncommonInitialismsReplacer = strings.NewReplacer(uncommonInitialismsForReplacer...)
}
