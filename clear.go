package main

import (
	"bufio"
	"io"
	logger "log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

func main() {
	//fmt.Printf("%#v", util.String2Slice("baidu"))
	run()
}

func run() {
	var (
		// 检测目录
		dir = "./"
		// 公司敏感词汇，利用字节数组过滤
		firmCode = []string{
			string([]byte{0x73, 0x6f, 0x6e, 0x67, 0x67, 0x75, 0x6f}),
			string([]byte{0x74, 0x74, 0x79, 0x63}),
			string([]byte{0x74, 0x74, 0x79, 0x6f, 0x6e, 0x67, 0x63, 0x68, 0x65}),
			string([]byte{0x70, 0x62, 0x73}),
		}
		// 所有待检测的文件
		allFile = make([]string, 0)
		err     error
		// 绝对路径
		absPath = AbsPath(dir)
	)
	Infof("开始检测代码: %+v, 位置: %s", ToCliMultiDescString(firmCode), absPath)

	if allFile, err = GetAllFiles(dir, func(fileName string) bool {
		dir := filepath.Dir(fileName)
		base := filepath.Base(fileName)
		if strings.HasPrefix(dir, ".idea") {
			return false
		}
		if strings.HasPrefix(dir, ".git") {
			return false
		}
		if strings.HasPrefix(dir, "bin") {
			return false
		}
		if base == ".DS_Store" {
			return false
		}
		return true
	}); err != nil {
		return
	}
	for _, elem := range allFile {
		Debugf("检测文件: %s", elem)
	}
	wg := sync.WaitGroup{}
	for _, elem := range allFile {
		wg.Add(1)
		go func(elem string) {
			file, err := os.Open(elem)
			defer func() {
				wg.Done()
				file.Close()
			}()
			if err != nil {
				panic(err)
			}
			if err := ReadFileLine(file, func(line string) error {
				for _, code := range firmCode {
					if strings.Contains(line, code) {
						Errorf("发现异常, 文件名称: %s, 检测出代码: %s", elem, code)
						panic("发现异常, 需要强制中断")
					}
				}
				return nil
			}); err != nil {
				panic(elem)
			}

		}(elem)
	}

	wg.Wait()

	Infof("完成检测代码: %+v, 位置: %s", ToCliMultiDescString(firmCode), absPath)
}

var (
	_log   = logger.New(os.Stdout, "[Clear] ", logger.LstdFlags)
	_warn  = "\033[33m[WARN]\033[0m "
	_error = "\033[31m[ERROR]\033[0m "
	_info  = "\033[32m[INFO]\033[0m "
	_debug = "\033[36m[DEBUG]\033[0m "

	Errorf = func(format string, v ...interface{}) {
		_log.Printf(_error+format, v...)
	}
	Warnf = func(format string, v ...interface{}) {
		_log.Printf(_warn+format, v...)
	}
	Infof = func(format string, v ...interface{}) {
		_log.Printf(_info+format, v...)
	}
	Debugf = func(format string, v ...interface{}) {
		_log.Printf(_debug+format, v...)
	}
)

func Slice2String(body []byte) string {
	if body == nil || len(body) == 0 {
		return ""
	}
	return *(*string)(unsafe.Pointer(&body))
}

type FilterFile func(fileName string) bool

func GetAllFiles(dirPth string, filter FilterFile) ([]string, error) {
	files := make([]string, 0)
	err := filepath.Walk(dirPth, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filter(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ReadFileLine(file io.Reader, foo func(line string) error) error {
	reader := bufio.NewReader(file)
	for {
		lines, isEOF, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if isEOF {
			break
		}
		if err := foo(string(lines)); err != nil {
			return err
		}
	}
	return nil
}

// 转换成 cli命令的 多个条件描述文本，例如 k1,k2 => "k1"|"k2"
func ToCliMultiDescString(slice []string) string {
	if slice == nil || len(slice) == 0 {
		return ""
	}
	lastIndex := len(slice) - 1
	result := strings.Builder{}
	for index, elem := range slice {
		result.WriteByte('"')
		result.WriteString(elem)
		result.WriteByte('"')
		if index != lastIndex {
			result.WriteByte('|')
		}
	}
	return result.String()
}

func AbsPath(filePath string) string {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		panic(err)
	}
	return absPath
}
