package main

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go/v7"
	"io"
	"minio-compress/utils"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup
var statisticsMap sync.Map
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func main() {
	utils.InitMinioClient()
	minioObjChan := utils.ListFiles()
	statisticsMap.Store("successNum", 0)
	statisticsMap.Store("errNum", 0)
	statisticsMap.Store("reduceSize", 0.00)
	semaphore := make(chan struct{}, 20)
	for objInfo := range minioObjChan {
		if objInfo.Err != nil {
			fmt.Println(objInfo.Err)
			continue
		}
		wg.Add(1)
		go func(obj minio.ObjectInfo) {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			semaphore <- struct{}{}
			file := obj.Key
			minioObj, err := utils.GetFile(file)
			if err != nil {
				fmt.Println("无法获取文件:", file)
				statisticsCount("errNum")
				return
			}
			defer minioObj.Close()
			stat, err := minioObj.Stat()
			if err != nil {
				fmt.Println("无法获取文件状态:", file)
				statisticsCount("errNum")
				return
			}
			contentType := stat.ContentType
			if !strings.Contains(contentType, "image") {
				fmt.Println("文件不是图片, 文件名:", file)
				statisticsCount("errNum")
				return
			}
			var filename, prefix string
			index := strings.LastIndex(file, "/")
			if index == -1 {
				filename = file
				prefix = ""
			} else {
				filename = file[index+1:]
				prefix = file[:index+1]
			}

			_ = os.MkdirAll("backupFiles/"+prefix, 0755)
			localFile, err := os.Create("backupFiles/" + prefix + filename)
			if err != nil {
				fmt.Println("文件保存失败, 文件名:", file, err)
				statisticsCount("errNum")
				return
			}
			defer localFile.Close()
			if _, err = io.Copy(localFile, minioObj); err != nil {
				os.Remove("backupFiles/" + prefix + filename) // 删除文件
				fmt.Println("文件拷贝失败, 文件名:", file, err)
				statisticsCount("errNum")
				return
			}

			compressFileBuffer := bufferPool.Get().(*bytes.Buffer)
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered in compressFileBuffer", r)
				}
				compressFileBuffer.Reset() // 使用完毕后重置
				bufferPool.Put(compressFileBuffer)
			}()
			err = utils.CompressFile("backupFiles/"+file, contentType, compressFileBuffer)
			if err != nil {
				fmt.Println("文件压缩失败, 文件名:", file, err)
				statisticsCount("errNum")
				return
			}
			compressedSize := compressFileBuffer.Len()
			if compressedSize >= int(stat.Size) || int(stat.Size)-compressedSize < 1024 {
				fmt.Println("文件过滤压缩, 压缩后体积无明显变化, 文件名:", file)
				statisticsCount("errNum")
				return
			}
			err = utils.UploadFile(file, contentType, compressFileBuffer)
			if err != nil {
				fmt.Println("文件压缩成功, 更新失败, 文件名:", file, err)
				statisticsCount("errNum")
			} else {
				statisticsSize("reduceSize", float64(int(stat.Size)-compressedSize))
				statisticsCount("successNum")
				fmt.Println("文件压缩成功,更新成功,文件名:", file, ",压缩前后体积:", stat.Size/1024, "KB /", compressedSize/1024, "KB")
			}

		}(objInfo)
	}
	wg.Wait()
	successNum, _ := statisticsMap.Load("successNum")
	errNum, _ := statisticsMap.Load("errNum")
	reduceSize, _ := statisticsMap.Load("reduceSize")
	fmt.Println("=================处理完成=================")
	fmt.Println("成功数量:", successNum)
	fmt.Println("失败数量:", errNum)
	fmt.Printf("总共减少体积: %.4f KB, 约 %.4f MB, 约 %.4f GB\n",
		reduceSize.(float64)/1024,
		reduceSize.(float64)/1024/1024,
		reduceSize.(float64)/1024/1024/1024)
}

func statisticsCount(key string) {
	v, _ := statisticsMap.Load(key)
	statisticsMap.Store(key, v.(int)+1)
}

func statisticsSize(key string, count float64) {
	v, _ := statisticsMap.Load(key)
	statisticsMap.Store(key, v.(float64)+count)
}
