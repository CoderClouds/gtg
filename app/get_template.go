/**
  create by yy on 2020/5/18
*/

package app

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

type GetTemplate struct {
	dirName string
}

func StartCreate(dirName string) {
	var (
		err error
	)

	getTemplate := &GetTemplate{dirName: dirName}

	getTemplate.replaceName()

	return
	if err = getTemplate.clone(); err != nil {
		err = NewReportError(err)
		fmt.Println(err)
	}
}

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

func (g *GetTemplate) replaceName() {
	// 首先获取所有文件列表
	var (
		err error
	)

	handleLockChan := make(chan int, 10)

	path := g.dirName

	if !strings.Contains(path, "/") {
		path = path + "/"
	}

	if err = g.traverseDir(path, handleLockChan); err != nil {
		err = NewReportError(err)
		fmt.Println(err)
	}

	for i := 0; i < 10; i++ {
		handleLockChan <- 1
	}

}

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

		go g.handleFile(tmpPath, handleLockChan)
	}

	return
}

func (g *GetTemplate) handleFile(path string, handleLockChan chan int) {
	fmt.Println(path)

	<-handleLockChan

	return
}
