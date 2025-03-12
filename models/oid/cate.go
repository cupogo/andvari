package oid

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/cupogo/andvari/models/idgen"
)

// ObjType 目标类型
type ObjType int16

// consts
const (
	OtDefault    ObjType = iota
	OtAccount            //  1 账号
	OtCompany            //  2 公司、企业、组织
	OtDepartment         //  3 部门
	OtArticle            //  4 内容、文章、条款
	OtTeam               //  5 小组、群
	OtEvent              //  6 事件：任务、消息、日志等
	OtToken              //  7 票据
	OtPeople             //  8 人员: 客户信息、联系人、地址等
	OtForm               //  9 表单: 配置、订单、票据等
	OtGoods              // 10 东西: 设备、配件、软件等
	OtFile               // 11 文件和文档等
	OtImage              // 12 图片
	OtLocale             // 13 位置
	OtMessage            // 14 消息
	OtProject            // 15 项目
	OtTask               // 16 任务
	otLast
)

func (ot ObjType) Code() string {
	switch ot {
	case OtDefault:
		return "de"
	case OtAccount:
		return "ac"
	case OtCompany:
		return "co"
	case OtDepartment:
		return "dp"
	case OtArticle:
		return "at"
	case OtTeam:
		return "tm"
	case OtToken:
		return "tk"
	case OtEvent:
		return "ev"
	case OtPeople:
		return "pe"
	case OtForm:
		return "fm"
	case OtGoods:
		return "go"
	case OtFile:
		return "fi"
	case OtImage:
		return "im"
	case OtLocale:
		return "lo"
	case OtMessage:
		return "ms"
	case OtProject:
		return "pj"
	case OtTask:
		return "ta"
	}
	return valCate(uint16(ot))
}

func ParseCate(s string) ObjType {
	switch s {
	case "ac", "account":
		return OtAccount
	case "co", "company":
		return OtCompany
	case "dp", "department":
		return OtDepartment
	case "at", "article":
		return OtArticle
	case "tm", "team":
		return OtTeam
	case "tk", "token":
		return OtToken
	case "ev", "event":
		return OtEvent
	case "pe", "people":
		return OtPeople
	case "fm", "form":
		return OtForm
	case "go", "goods":
		return OtGoods
	case "fi", "file":
		return OtFile
	case "im", "image":
		return OtImage
	case "lo", "locale":
		return OtLocale
	case "ms", "message":
		return OtMessage
	case "pj", "project":
		return OtProject
	case "ta", "task":
		return OtTask
	default:
		return OtDefault
	}
}

var (
	longNames = map[string]string{
		"department": "dp",
		"article":    "at",
		"team":       "tm",
		"token":      "tk",
		"form":       "fm",
		"message":    "ms",
		"project":    "pj",
	}
	prefixes = map[string]uint16{
		"ac": 1,
		"co": 2,
		"dp": 3,
		"at": 4,
		"tm": 5,
		"tk": 6,
		"ev": 7,
		"pe": 8,
		"fm": 9,
		"go": 10,
		"fi": 11,
		"im": 12,
		"lo": 13,
		"ms": 14,
		"pj": 15,
		"ta": 16,
	}
	cateLock sync.Mutex
)

const (
	minCate uint16 = 360 // a0 == 360
)

func cateVal(s string) uint16 {
	if len(s) == 0 {
		panic("empty code")
	}
	if s[0] < 'a' || s[0] == ' ' {
		panic(fmt.Errorf("invalid code: %s", s))
	}
	if len(s) == 1 {
		s = s + "a"
	} else if s[1] < 'a' || s[1] == ' ' {
		panic(fmt.Errorf("invalid code: %s", s))
	}

	var bI big.Int
	i, _ := bI.SetString(s[0:2], 36)
	return uint16(i.Uint64()) - minCate
}

func valCate(n uint16) string {
	var bI big.Int
	i := bI.SetUint64(uint64(n + minCate))
	return i.Text(36)
}

func RegistCate(name, code string) {
	if len(name) < 2 {
		panic(fmt.Errorf("too shart name: %s", name))
	}
	if len(code) < 2 {
		panic(fmt.Errorf("too short code: %s", code))
	}
	if len(code) > 2 {
		code = code[0:2]
	}
	code = strings.ToLower(code)
	name = strings.ToLower(name)

	if ParseCate(name) != OtDefault {
		panic(fmt.Errorf("exist name %s", name))
	}

	if ParseCate(code) != OtDefault {
		panic(fmt.Errorf("exist code %s", code))
	}

	cateLock.Lock()
	defer cateLock.Unlock()

	if _, ok := prefixes[code]; ok {
		panic(fmt.Errorf("exist code %s", code))
	}

	if name[0:2] != code {
		if _, ok := longNames[name]; ok {
			panic(fmt.Errorf("exist name %s", name))
		}
		longNames[name] = code
	}
	sm := cateVal(code)
	prefixes[code] = sm
	shards[ObjType(sm)] = idgen.NewWithShard(int64(sm))
}

func NewWithCode(code string) (OID, bool) {
	code = strings.ToLower(code)
	if ot := ParseCate(code); ot != OtDefault {
		return NewID(ot), true
	}
	if s, ok := longNames[code]; ok {
		code = s
	} else if len(code) > 2 {
		code = code[0:2]
	}
	if val, ok := prefixes[code]; ok {
		return NewID(ObjType(val)), ok
	}
	return ZeroID, false
}
