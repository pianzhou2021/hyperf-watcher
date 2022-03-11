package FSNotify

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type NotifyFile struct {
	watch  *fsnotify.Watcher
	reload bool
}

func NewNotifyFile() *NotifyFile {
	w := new(NotifyFile)
	w.watch, _ = fsnotify.NewWatcher()
	return w
}

//监控目录
func (ntf *NotifyFile) WatchDir(dir string, exclude string) {
	//通过Walk来遍历目录下的所有子目录

	skipDir := strings.Split(exclude, ",")
	// fmt.Println("skip dir:", skipDir)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		//判断是否为目录，监控目录,目录下文件也在监控范围内，不需要加
		if info.IsDir() {
			path, err := filepath.Abs(path)

			name := strings.Replace(path, dir+"/", "", -1)
			if InArray(skipDir, name) {
				fmt.Println("skip dir: ", path)
				return filepath.SkipDir
			}

			if err != nil {
				return err
			}

			err = ntf.watch.Add(path)
			if err != nil {
				return err
			}

			fmt.Println("Watch : ", path)
		}
		return nil
	})

	go ntf.WatchEvent() //协程
}

func (ntf *NotifyFile) WatchEvent() {
	for {
		select {
		case ev := <-ntf.watch.Events:
			{
				if ev.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println("Create File : ", ev.Name)
					//获取新创建文件的信息，如果是目录，则加入监控中
					file, err := os.Stat(ev.Name)
					if err == nil && file.IsDir() {
						ntf.watch.Add(ev.Name)
						fmt.Println("Add Watch : ", ev.Name)
						ntf.reload = true
					}
				}

				if ev.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("Write File : ", ev.Name)
					ntf.reload = true
				}

				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					fmt.Println("Remove File : ", ev.Name)
					//如果删除文件是目录，则移除监控
					fi, err := os.Stat(ev.Name)
					if err == nil && fi.IsDir() {
						ntf.watch.Remove(ev.Name)
						fmt.Println("Remove Watch Dir : ", ev.Name)
					}
					ntf.reload = true
				}

				if ev.Op&fsnotify.Rename == fsnotify.Rename {
					//如果重命名文件是目录，则移除监控 ,注意这里无法使用os.Stat来判断是否是目录了
					//因为重命名后，go已经无法找到原文件来获取信息了,所以简单粗爆直接remove
					fmt.Println("Rename File : ", ev.Name)
					ntf.watch.Remove(ev.Name)
					ntf.reload = true
				}

				if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
					fmt.Println("Modify File : ", ev.Name)
					ntf.reload = true
				}
			}
		case err := <-ntf.watch.Errors:
			{
				fmt.Println("error : ", err)
				return
			}
		}
	}
}

func (ntf *NotifyFile) Command(command string) {
	if ntf.reload {
		fmt.Println("Command: ", command)

		go Command(command)

		ntf.reload = false
	}
}

func Command(command string) {
	names := strings.Split(command, " ")
	cmd := exec.Command(names[0], (names[1:])...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println("Result: ", out.String())
}

func InArray(arr []string, target interface{}) bool {

	for _, element := range arr {
		element = strings.Trim(element, " ")
		if element == target {
			return true
		}
	}

	return false
}
