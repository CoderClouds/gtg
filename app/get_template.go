/**
  create by yy on 2020/5/18
*/

package app

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type GetTemplate struct {
	dirName    string
	replaceStr string
	finalName  string
}

// 启动程序(入口程序)
func StartCreate(dirName string) {
	var (
		err error
	)

	getTemplate := &GetTemplate{
		dirName:    dirName,
		replaceStr: "gin_template",
	}

	if err = getTemplate.clone(); err != nil {
		err = NewReportError(err)
		fmt.Println(err)
	}

	getTemplate.replaceName()
}

// 从git仓库拉取 模板文件 并且命名为 用户自定义的名称 dirName
func (g *GetTemplate) clone() (err error) {

	var (
		bytes []byte
	)

	// 这里应该做一个容错操作，如果当前目录下已经有了gin_template文件夹，
	// 则先将 原本的 gin_template 文件夹进行改名，在 clone的gin_template文件夹被改名之后再恢复其名称
	// 为了防止 改名操作出现重名， 用 MD5 加密后的字符串 作为文件名
	command := "git clone https://github.com/guaidashu/gin_template.git"
	fmt.Println(command)

	cmd := exec.Command("/bin/bash", "-c", command)

	if bytes, err = cmd.Output(); err != nil {
		fmt.Println(string(bytes))
		err = NewReportError(err)
		log.Println(err)
		return
	}

	command2 := "mv gin_template %v"
	command2 = fmt.Sprintf(command2, g.dirName)
	fmt.Println(command2)

	cmd2 := exec.Command("/bin/bash", "-c", command2)

	if bytes, err = cmd2.Output(); err != nil {
		err = NewReportError(err)
		log.Println(err)
		if err = g.deleteTmpDir(); err != nil {
			fmt.Println(err)
		}
		return
	}

	command3 := "cd %v && rm -rf .idea && rm -rf .git"
	command3 = fmt.Sprintf(command3, g.dirName)
	cmd3 := exec.Command("/bin/bash", "-c", command3)

	if bytes, err = cmd3.Output(); err != nil {
		err = NewReportError(err)
		log.Println(err)
		return
	}

	// 接下来进行文件替换
	return
}

// 替换所有 gin_template 字样
func (g *GetTemplate) replaceName() {
	// 首先获取所有文件列表
	var (
		err error
	)

	handleLockChan := make(chan int, 10)

	path := g.dirName

	nameArr := strings.Split(g.dirName, "/")
	g.finalName = nameArr[len(nameArr)-1]
	if g.finalName == "" {
		g.finalName = nameArr[len(nameArr)-2]
	}

	if !strings.Contains(path, "/") {
		path = path + "/"
	}

	if err = g.traverseDir(path, handleLockChan); err != nil {
		err = NewReportError(err)
		fmt.Println(err)
	}

	// 阻塞，主要是有递归存在，所以不能直接用一个channel来进行阻塞，这里选择填充满
	// handleLockChan来 判断程序是否执行完，如果是满的则 会阻塞，
	// 一直到所有协程执行完毕，才能塞入值解除阻塞
	for i := 0; i < 10; i++ {
		handleLockChan <- 1
	}

}

// clone出错或者 其他操作出错的时候 删除文件夹
func (g *GetTemplate) deleteTmpDir() (err error) {

	var (
		bytes []byte
	)

	command := "rm -rf %v"
	command = fmt.Sprintf(command, g.dirName)

	cmd := exec.Command("/bin/bash", "-c", command)

	if bytes, err = cmd.Output(); err != nil {
		err = NewReportError(err)
		log.Println(err)
		fmt.Println(string(bytes))
	}

	return
}

// 遍历所有文件
func (g *GetTemplate) traverseDir(path string, handleLockChan chan int) (err error) {
	var (
		ok      bool
		tmpPath string
	)

	// 对path进行处理，

	files, _ := ioutil.ReadDir(path)

	for _, f := range files {
		tmpPath = path + f.Name()
		// 如果是文件夹，则进行递归遍历 扫描文件夹
		if ok = IsDir(tmpPath); ok {
			tmpPath = tmpPath + "/"
			if err = g.traverseDir(tmpPath, handleLockChan); err != nil {
				err = NewReportError(err)
				fmt.Println(err)
				return
			}
			continue
		}

		// 如果是文件，则直接读取并 进行相关的内容修改操作
		// 以协程的方式打开，为了加速效率，但是需要限制一定量，目前 限制为同时打开十个文件
		// 所以需要创建一个 限制的channel

		handleLockChan <- 1

		go g.handleFile(path, f.Name(), handleLockChan)
	}

	return
}

// 读写文件
// 此函数的读写规则是边读边写，牺牲内存提高效率
func (g *GetTemplate) handleFile(path, fileName string, handleLockChan chan int) {
	var (
		err       error
		readFile  *os.File
		writeFile *os.File
		readSize  int
		bufSize   int
	)

	// 首先读文件
	oldPath := path + fileName
	readFile, err = os.Open(oldPath)

	// 打开 要一个临时文件，用于写入
	newPath := path + "new_" + fileName
	writeFile, err = os.OpenFile(newPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	defer func() {
		if err = readFile.Close(); err != nil {
			fmt.Println(err)
		}
		if err = writeFile.Close(); err != nil {
			fmt.Println(err)
		}

		// 删除源文件
		if err = os.Remove(oldPath); err != nil {
			fmt.Println(err)
		}
		// 移动新文件为源文件名
		if err = os.Rename(newPath, oldPath); err != nil {
			fmt.Println(err)
		}

		<-handleLockChan
	}()

	reader := bufio.NewReader(readFile)
	bufSize = 512
	buf := make([]byte, bufSize)

	for {
		// 读数据
		readSize, err = reader.Read(buf)
		s := string(buf[:readSize])

		// 写入
		if _, err = writeFile.Write([]byte(strings.Replace(s, g.replaceStr, g.finalName, -1))); err != nil {
			fmt.Println(err)
		}

		if readSize < bufSize || readSize <= 0 || err == io.EOF {
			break
		}
	}
}
