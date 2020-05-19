/**
  create by yy on 2020/5/18
*/

package app

import (
	"fmt"
	"log"
	"os/exec"
)

type GetTemplate struct {
	dirName string
}

func StartCreate(dirName string) {
	getTemplate := &GetTemplate{dirName: dirName}

	getTemplate.clone()
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
		log.Println(err)
		return
	}

	command2 := "mv gin_template %v"
	command2 = fmt.Sprintf(command2, g.dirName)
	fmt.Println(command2)

	cmd2 := exec.Command("/bin/bash", "-c", command2)

	if bytes, err = cmd2.Output(); err != nil {
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
		log.Println(err)
		return
	}

	// 接下来进行文件替换
	return
}

func (g *GetTemplate) replaceName() {
	// 首先获取所有文件列表

}

func (g *GetTemplate) deleteTmpDir() (err error) {

	var (
		bytes []byte
	)

	command := "rm -rf %v"
	command = fmt.Sprintf(command, g.dirName)

	cmd := exec.Command("/bin/bash", "-c", command)

	if bytes, err = cmd.Output(); err != nil {
		log.Println(err)
		fmt.Println(string(bytes))
	}

	return
}
