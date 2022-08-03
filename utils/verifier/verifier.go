package verifier

import (
	"regexp"
	"strconv"
	"strings"
	"github.com/c-jimin/codetech/utils/config"
)

var (
	fileSystemAllowTypeDict map[string]interface{}
)

func init() {
	fileSystemAllowTypeDict = make(map[string]interface{}, 0)
	for _, v := range config.FileSystemAllowTypeList {
		ss := strings.Split(v, "/")
		if len(ss) != 2 {
			continue
		}
		k, ok := fileSystemAllowTypeDict[ss[0]]
		if !ok {
			if ss[1] == "*" {
				fileSystemAllowTypeDict[ss[0]] = "*"
			} else {
				fileSystemAllowTypeDict[ss[0]] = []string{ss[1]}
			}
		} else {
			if ss[1] == "*" {
				fileSystemAllowTypeDict[ss[0]] = "*"
			} else {
				switch k.(type) {
				case string:
					continue
				case []string:
					tmp := fileSystemAllowTypeDict[ss[0]].([]string)
					tmp = append(tmp, ss[1])
				}
			}
		}
	}
}
func IsEmail(email string) bool {
	return regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`).MatchString(email)
}

func IsPhone(phone string) bool {
	return regexp.MustCompile(`^1[34578]\d{9}$`).MatchString(phone)
}

func IsInt(string string) bool {
	_, err := strconv.Atoi(string)
	if err != nil {
		return false
	}
	return true
}

func IsItInAllowTypeList(typ string) bool {
	ss := strings.Split(typ, "/")
	if len(ss) != 2 {
		return false
	}
	k, ok := fileSystemAllowTypeDict[ss[0]]
	if !ok {
		return false
	} else {
		switch k.(type) {
		case string:
			return true
		case []string:
			tmp := k.([]string)
			for _, v := range tmp {
				if v == ss[1] {
					return true
				}
			}
			return false
		default:
			return false
		}
	}
}
