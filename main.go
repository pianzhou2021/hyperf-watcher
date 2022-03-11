/*
 * @Description:
 * @Author: (c) Pian Zhou <pianzhou2021@163.com>
 * @Date: 2022-03-04 17:25:03
 * @LastEditors: Pian Zhou
 * @LastEditTime: 2022-03-10 19:31:48
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"watcher/FSNotify"
)

func main() {
	path, _ := os.Executable()
	include := filepath.Dir(filepath.Dir(path))
	//获取命令参数
	exclude := flag.String("exclude", ".git, vendor, runtime", "excluded files or directories")
	command := flag.String("command", "php "+include+"/bin/hyperf.php serve:restart", "the command after change")
	ttl := flag.Int("ttl", 1000, "frequency of check (ms)")
	// 解析命令行参数写入注册的flag里
	flag.Parse()
	fmt.Println("")
	fmt.Println("                    Hyperf Watcher By PianZhou (pianzhou2021@163.com)                        ")
	fmt.Println("")
	fmt.Println("--exclude :", *exclude)
	fmt.Println("--command :", *command)
	fmt.Println("--ttl :", *ttl)

	watch := FSNotify.NewNotifyFile()
	watch.WatchDir(include, *exclude)

	//定时启动
	for range time.Tick(time.Duration(*ttl) * time.Millisecond) {
		// fmt.Printf("%d 毫秒定时检测重载\r\n", *ttl)
		watch.Command(*command)
	}

	select {}
}
