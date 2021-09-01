// Copyright © 2018 Alex Goodman
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	_ "github.com/go-sql-driver/mysql"
	"github.com/wagoodman/dive/cmd"
	"github.com/wagoodman/dive/dive"
	"github.com/wagoodman/dive/dive/image"
)

var (
	version   = "No version provided"
	commit    = "No commit provided"
	buildTime = "No build timestamp provided"
)

func checkErr(err error) {
	if err != nil {
		a += 1
		fmt.Println(err)
		//panic(err)

	}
}

var a int

//阻塞式的执行外部shell命令的函数,等待执行完毕并返回标准输出
func exec_shell(s string) (string, error) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command("docker", "manifest", "inspect", s)

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out

	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	err := cmd.Run()
	checkErr(err)

	return out.String(), err
}

func delete_image(imagename string) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	checkErr(err)

	images, err := cli.ImageRemove(ctx, imagename, types.ImageRemoveOptions{})
	// images, err := cli.DistributionInspect(ctx, imagename, "")
	// images, err := cli.ImageList(ctx, types.ImageListOptions{})

	checkErr(err)

	fmt.Println("already delete ", imagename, "len : ", len(images))
	// config, bytes, _ := cli.ConfigInspectWithRaw(ctx, string(images.Descriptor.Digest))
	// fmt.Println("on : ", config)
	// fmt.Println("on : ", bytes)

	// for _, image := range images {
	// 	fmt.Println(image.Deleted)
	// 	// fmt.Println(image.RepoTags)
	// }
}

func queryData(Db *sql.DB, num1 int) string {
	var get_name string
	rows, err := Db.Query("select * from image_id_list limit " + strconv.Itoa(num1) + ",1")
	checkErr(err)

	for rows.Next() {
		//定义变量接收查询数据
		var name string
		err := rows.Scan(&name)
		checkErr(err)
		get_name = name
	}
	//关闭结果集（释放连接）
	rows.Close()
	return get_name
}

func get_image_detail(image_id string) {
	var sourceType dive.ImageSource
	var imageStr string
	userImage := image_id
	sourceType, imageStr = dive.DeriveImageSource(userImage)
	if sourceType == dive.SourceUnknown {
		sourceStr := "docker"
		sourceType = dive.ParseImageSource(sourceStr)
		if sourceType == dive.SourceUnknown {
			fmt.Printf("unable to determine image source: %v\n", sourceStr)
			os.Exit(1)
		}
		imageStr = userImage
	}
	imageResolver, _ := dive.GetImageResolver(sourceType)

	var img *image.Image
	img, _ = imageResolver.Fetch(imageStr)
	analysis, _ := img.Analyze()
	fmt.Println("analyse over ! layer len : ", len(analysis.Layers))
	// for i, x := range analysis.Layers {
	// 	fmt.Println(i, "\n", x.Id, "\n", x.Digest)
	// 	// return x.Id, x.Digest
	// }

}

func get_image_id(num int) string {
	conn, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/test")
	if nil != err {
		fmt.Println("connect db error: ", err)
	}
	get_name := queryData(conn, num)
	print(get_name)
	return get_name

}

func main() {
	a = 0
	for i := 650; i < 1000; i++ {
		name := get_image_id(i)
		// get_image_detail(name)
		fmt.Println(exec_shell(name))

		fmt.Println(i, "---err:", a, "-----", name)
		// delete_image(name)
		// fmt.Println(id)
		// fmt.Println(digest)
	}
	return

	cmd.SetVersion(&cmd.Version{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	})

	cmd.Execute()
}
