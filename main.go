/**
  create by yy on 2019-10-21
*/

package main

import (
	"flag"
)

func main() {
	InitFlag()
	flag.Parse()
	if h {
		flag.Usage()
	}
}
