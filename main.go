/**
  create by yy on 2019-10-21
*/

package main

import (
	"fmt"
	"os"
)

func main() {
	currentPath, _ := os.Getwd()
	fmt.Println(currentPath)
}