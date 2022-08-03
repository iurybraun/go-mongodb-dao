package exception

import (
	"fmt"
	"github.com/c-jimin/codetech/utils/config"
	"runtime"
	"strconv"
	"strings"
)

type Exception struct {
	SimpleException
	err   *error
	trace *trace
}

type SimpleException struct {
	code uint16
	key  string
	info string
}
type trace struct {
	pc   uintptr
	file string
	line int
}

var (
	BadOperate              = &SimpleException{40002, "BadOperate", "Bad Operate"}
	UnknownError            = &SimpleException{40010, "UnknownError", "未知错误"}
	MissingParameters       = &SimpleException{40011, "MissingParameters", "缺少参数"}
	UIDFormatError          = &SimpleException{40012, "UIDFormatError", "UID格式错误"}
	RoleError               = &SimpleException{40013, "RoleError", "用户角色错误"}
	UsernameFormatError     = &SimpleException{40014, "UsernameFormatError", "用户名格式错误"}
	PasswordFormatError     = &SimpleException{40015, "PasswordFormatError", "密码格式错误"}
	EmailFormatError        = &SimpleException{40016, "EmailFormatError", "邮箱格式错误"}
	SexFormatError          = &SimpleException{40017, "SexError", "性别格式错误"}
	PhoneNumberFormatError  = &SimpleException{40018, "PhoneNumberFormatError", "手机号码格式错误"}
	BirthdayError           = &SimpleException{40019, "BirthdayError", "生日日期错误"}
	BlockLevelError         = &SimpleException{40020, "BlockLevelError", "封禁等级错误"}
	DescriptionFormatError  = &SimpleException{40021, "DescriptionFormatError", "自我描述格式错误"}
	IntroductionFormatError = &SimpleException{40022, "IntroductionFormatError", "个人介绍格式错误"}
	ObjectIDFormatError     = &SimpleException{40023, "ObjectIDFormatError", "ObjectID格式错误"}
	DataFormatError         = &SimpleException{40024, "DataFormatError", "数据格式化错误"}

	Unauthorized = &SimpleException{40101, "Unauthorized", "用户身份未验证"}
	LoginFailed  = &SimpleException{40102, "LoginFailed", "用户名或密码不正确"}
	NoPermission = &SimpleException{40103, "NoPermission", "无权限执行该操作"}

	NotOnChangeUsernameDate = &SimpleException{40301, "NotOnChangeUsernameDate", "未到更改用户名日期"}
	NotOnChangeSexDate      = &SimpleException{40302, "NotOnChangeSexDate", "未到更改性别日期"}
	NotOnCheckInTime        = &SimpleException{40303, "NotOnCheckInTime", "未到签到时间"}
	InsufficientBalance     = &SimpleException{40304, "InsufficientBalance", "余额不足"}
	EmailAlreadyExist       = &SimpleException{40305, "EmailAlreadyExist", "邮箱已存在"}
	UsernameAlreadyExist    = &SimpleException{40306, "UsernameAlreadyExist", "用户名已存在"}
	PhoneNumberAlreadyExist = &SimpleException{40307, "PhoneNumberAlreadyExist", "手机号已存在"}
	QuestionIsClosed        = &SimpleException{40308, "QuestionIsClosed", "问题已关闭"}
	AnswerAlreadyExist      = &SimpleException{40309, "AnswerAlreadyExist", "已经回答过该问题"}
	TooManyLabel            = &SimpleException{40310, "TooManyLabel", "标签数量达到上限"}
	LabelIsBanned           = &SimpleException{40311, "LabelIsBanned", "标签被禁用"}
	UserIsBanned            = &SimpleException{40312, "UserIsBanned", "当前处于封禁状态"}
	CantBeEmpty             = &SimpleException{40313, "CantBeEmpty", "不能为空"}
	AlreadyGrade            = &SimpleException{40314, "AlreadyGrade", "你已经评过分了"}
	AlreadyAnswered         = &SimpleException{40315, "AlreadyAnswered", "你已经回答过该问题了"}
	OldPasswordError        = &SimpleException{40316, "OldPasswordError", "原密码错误"}

	RecordNotFound  = &SimpleException{40401, "RecordNotFound", "未找到该记录"}
	RecordIsDeleted = &SimpleException{40402, "RecordIsDeleted", "记录被删除"}
	UserNotOnline   = &SimpleException{40403, "UserNotOnline", "该用户当前不在线"}

	TokenTimeout = &SimpleException{40800, "Token Timeout", "token已过期"}
)

func (se *SimpleException) New(v ...interface{}) *Exception {
	except := new(Exception)
	except.code = se.code
	except.key = se.key
	except.info = se.info
	for _, v := range v {
		switch typ := v.(type) {
		case error:
			except.err = &typ
		case string:
			except.info = typ
		}
	}
	funcName, file, line, ok := runtime.Caller(1)
	if ok {
		except.trace = &trace{funcName, file, line}
	}
	return except
}
func (except *Exception) NewInfo(info string) *Exception {
	if except == nil {
		return except
	}
	except.info = info
	return except
}
func (except *Exception) NewInfoVerifyKey(key string, info string) *Exception {
	if except == nil {
		return except
	}
	if except.key == key {
		except.info = info
	}
	return except
}
func (except *Exception) GetCode() int {
	if except == nil {
		return 20000
	}
	return int(except.code)
}
func (except *Exception) GetHttpCode() int {
	if except == nil {
		return 200
	}
	httpCode, _ := strconv.Atoi(strconv.Itoa(int(except.code))[:3])
	return httpCode
}
func (except *Exception) GetInfo() string {
	if except == nil {
		return "无异常"
	}
	return except.info
}
func (except *Exception) CheckKey(key string) bool {
	if except == nil {
		return false
	}
	return except.key == key
}
func (except *Exception) CheckError(err error) bool {
	if except == nil || except.err == nil {
		return false
	}
	return (*except.err).Error() == err.Error()
}

func (except *Exception) String() string {
	if except == nil {
		return "无异常"
	}
	if config.Debug == true {
		info := []string{
			except.info,
		}
		if except.err != nil {
			info = append(info, (*except.err).Error())
		}
		return strings.Join(info, "\n")
	} else {
		return except.info
	}
}
func (except *Exception) Format() string {
	if except == nil {
		return "无异常"
	}
	info := []string{
		except.info,
	}
	if except.err != nil {
		info = append(info, (*except.err).Error())
	}
	return strings.Join(info, "\n")
}
func (except *Exception) Trace() string {
	if except == nil {
		return "无追踪信息"
	}
	var info []string
	if except.trace != nil {
		info = append(info, fmt.Sprintf("%s 的追踪信息：", except.key))
		info = append(info, fmt.Sprintf("%s", runtime.FuncForPC(except.trace.pc).Name()))
		info = append(info, fmt.Sprintf("%s,line: %d", except.trace.file, except.trace.line))
	} else {
		info = append(info, fmt.Sprintf("%s 不含有追踪信息。", except.key))
	}
	return strings.Join(info, "\n")
}

func (except *Exception) Error() string {
	if except == nil {
		return "无异常"
	}

	info := []string{
		except.info,
	}
	if except.err != nil {
		info = append(info, (*except.err).Error())
	}
	return strings.Join(info, "\n")
}
