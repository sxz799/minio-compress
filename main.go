package main

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go/v7"
	"io"
	"math"
	"minio-compress/utils"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup
var mu sync.Mutex
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func main() {
	utils.InitMinioClient()

	minioObjChan := utils.ListFiles("2024/01/01/")
	var successNum, errNum, reduceSize int
	for objInfo := range minioObjChan {
		if objInfo.Err != nil {
			fmt.Println(objInfo.Err)
			continue
		}
		wg.Add(1)
		go func(obj minio.ObjectInfo) {
			defer wg.Done()
			file := objInfo.Key
			minioObj, err := utils.GetFile(file)
			if err != nil {
				handleError("无法获取文件:", file, err, &mu, &errNum)
				return
			}
			defer minioObj.Close()
			stat, err := minioObj.Stat()
			if err != nil {
				handleError("无法获取文件状态:", file, err, &mu, &errNum)
				return
			}
			contentType := stat.ContentType
			if !strings.Contains(contentType, "image") {
				handleError("文件不是图片, 文件名:", file, nil, &mu, &errNum)
				return
			}
			index := strings.LastIndex(file, "/")
			filename := file[index+1:]
			prefix := file[:index+1]
			_ = os.MkdirAll(prefix, 0755)
			localFile, err := os.Create(prefix + filename)
			if err != nil {
				handleError("文件保存失败, 文件名:", file, err, &mu, &errNum)
				return
			}
			defer localFile.Close()
			if _, err = io.Copy(localFile, minioObj); err != nil {
				handleError("文件拷贝失败, 文件名:", file, err, &mu, &errNum)
				return
			}

			compressFileBuffer := bufferPool.Get().(*bytes.Buffer)
			compressFileBuffer.Reset()
			defer func() {
				compressFileBuffer.Reset() // 使用完毕后重置
				bufferPool.Put(compressFileBuffer)
			}()
			err = utils.CompressFile(file, contentType, compressFileBuffer)
			if err != nil {
				handleError("文件压缩失败, 文件名:", file, err, &mu, &errNum)
				return
			}
			compressedSize := compressFileBuffer.Len()
			if math.Abs(float64(compressedSize)-float64(stat.Size)) < 1024 {
				handleError("压缩失败, 压缩后体积与压缩前体积不超过1KB, 文件名:", file, nil, &mu, &errNum)
				return
			}
			if compressedSize >= int(stat.Size) {
				handleError("文件压缩失败, 压缩后体积大于压缩前体积, 文件名:", file, nil, &mu, &errNum)
				return
			}
			err = utils.UploadFile(file, contentType, compressFileBuffer)
			if err != nil {
				handleError("文件压缩成功, 更新失败, 文件名:", file, err, &mu, &errNum)
			} else {
				mu.Lock()
				reduceSize += int(stat.Size) - compressedSize
				successNum++
				mu.Unlock()
				fmt.Println("文件压缩成功,更新成功,文件名:", file, ",压缩前后体积:", stat.Size/1024, "KB /", compressedSize/1024, "KB")
			}

		}(objInfo)
	}
	wg.Wait()
	fmt.Println("=================处理完成=================")
	fmt.Println("成功数量:", successNum)
	fmt.Println("失败数量:", errNum)
	fmt.Println("总共减少体积:", reduceSize/1024, "KB", ",约", reduceSize/1024/1024, "MB", ",约", reduceSize/1024/1024/1024, "GB")
}

func handleError(msg string, file string, err error, mu *sync.Mutex, errNum *int) {
	fmt.Println(msg, file, "error:", err)
	mu.Lock()
	*errNum++
	mu.Unlock()
}
