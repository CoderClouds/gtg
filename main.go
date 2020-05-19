/**
  create by yy on 2019-10-21
*/

package main

import (
	"flag"
	"fmt"
	"gtg/app"
	"os"
)

func main() {
	// InitFlag()
	var (
		flagArr []string
	)

	// flag.StringVar(&dirName, "name", "", "use gin_template to create a new project")

	flag.Parse()

	flagArr = flag.Args()

	// if h {
	// 	flag.Usage()
	// }

	if len(flagArr) >= 1 {
		// 进入选择
		switch flagArr[0] {
		case "new":
			// 进入 创建环节
			// 首先判断是否有文件夹名字
			if len(flagArr) > 1 {
				app.StartCreate(flagArr[1])
			} else {
				fmt.Println("Please input a director name.")
				os.Exit(1)
			}
		default:
			fmt.Println("Please input a correct command")

		}
	}

	// fmt.Println(flagArr)
}
