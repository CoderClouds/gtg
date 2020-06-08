/**
  create by yy on 2020/5/18
*/

package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type stdType string

const (
	STDERR stdType = "stderr"
	STDOUT stdType = "stdout"
)

type GetTemplate struct {
	dirName    string
	replaceStr string
	finalName  string
	isExists   bool
	cloneAddr  string
}

// 启动程序(入口程序)
func StartCreate(dirName string) {
	var (
		err error
	)

	getTemplate := &GetTemplate{
		dirName:    dirName,
		replaceStr: "gin_template",
		isExists:   false,
		cloneAddr:  "https://github.com/guaidashu/gin_template.git",
	}

	defer func() {
		if getTemplate.isExists {
			// 如果是移动了 原来的文件夹，则还原
			_, _ = getTemplate.executeScript(fmt.Sprintf("mv %v_old %v", getTemplate.replaceStr, getTemplate.replaceStr), STDERR)
		}
	}()

	if err = getTemplate.clone(); err != nil {
		err = NewReportError(err)
		fmt.Println(err)
		return
	}

	getTemplate.replaceName()

}

// 从git仓库拉取 模板文件 并且命名为 用户自定义的名称 dirName
func (g *GetTemplate) clone() (err error) {
	// 这里应该做一个容错操作，如果当前目录下已经有了gin_template文件夹，
	// 则先将 原本的 gin_template 文件夹进行改名，在 clone的gin_template文件夹被改名之后再恢复其名称
	// 为了防止 改名操作出现重名， 用 MD5 加密后的字符串 作为文件名

	// 开始判断是否存在gin_template
	var (
		result string
	)

	if result, err = g.executeScript("ls", STDOUT, false); err != nil {
		err = NewReportError(err)
		return
	}

	if strings.Contains(result, fmt.Sprintf("%v", g.replaceStr)) {
		// 如果存在就进行移动
		g.isExists = true
		_, _ = g.executeScript(fmt.Sprintf("mv %v %v_old", g.replaceStr, g.replaceStr), STDERR)
	}

	// 拉取源项目
	if _, err = g.executeScript(fmt.Sprintf("git clone %v", g.cloneAddr), STDERR); err != nil {
		err = NewReportError(err)
		return
	}

	// 将clone文件命名为 目标文件夹名
	command2 := "mv %v %v"
	command2 = fmt.Sprintf(command2, g.replaceStr, g.dirName)
	fmt.Println(command2)

	if _, err = g.executeScript(command2, STDERR); err != nil {
		err = NewReportError(err)
		log.Println(err)
		if err = g.deleteTmpDir(); err != nil {
			err = NewReportError(err)
		}
		return
	}

	// 进入 项目文件夹删除.idea 文件夹和 .git文件夹
	command3 := "cd %v && rm -rf .idea && rm -rf .git"
	command3 = fmt.Sprintf(command3, g.dirName)
	fmt.Println(command3)
	if _, err = g.executeScript(command3, STDERR); err != nil {
		err = NewReportError(err)
		return
	}

	return
}

// 通用执行脚本函数(默认会打印内容, 传入一个参数则进行判断)
func (g *GetTemplate) executeScript(command string, mode stdType, isPrintArr ...bool) (result string, err error) {
	var (
		ctx     context.Context
		scanner *bufio.Scanner
		message string
		stdErr  io.ReadCloser
		isPrint = true
	)

	if len(isPrintArr) > 0 {
		isPrint = isPrintArr[0]
	}

	ctx = context.TODO()

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", command)

	switch mode {
	case STDERR:
		if stdErr, err = cmd.StderrPipe(); err != nil {
			err = NewReportError(err)
			return
		}
	case STDOUT:
		if stdErr, err = cmd.StdoutPipe(); err != nil {
			err = NewReportError(err)
			return
		}
	}

	if err = cmd.Start(); err != nil {
		err = NewReportError(err)
		return
	}

	scanner = bufio.NewScanner(stdErr)
	scanner.Split(bufio.ScanLines)

	if isPrint {
		for scanner.Scan() {
			message = scanner.Text()
			result = result + message
			if _, err = fmt.Fprintf(os.Stderr, "%v\n", message); err != nil {
				err = NewReportError(err)
				return
			}
		}
	}

	if err = cmd.Wait(); err != nil {
		err = NewReportError(err)
	}

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
