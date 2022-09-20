package main

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"os"
	"path"
	"strconv"
)

func MD5(str string) string {
	data := []byte(str) //切片
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}

type UploadFile struct {
	Id             int `gorm:"primaryKey;AUTO_INCREMENT"`
	FileName       string
	FileSize       float64
	Hash           string
	DownloadNumber int
	DownloadPath   string
}

var (
	// 创建一个全局变量
	db        *gorm.DB
	ACCESSKEY = "qwert@905008"
)

func Initialize() (*gorm.DB, error) {
	var err error
	db, err = gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
		return db, err
	} else {
		// Migrate the schema
		//db.AutoMigrate(&Product{})

		return db, err
	}

}

func main() {
	db, ormerror := Initialize()
	if ormerror != nil {
		panic(ormerror)
	}
	migrateErr := db.AutoMigrate(&UploadFile{})
	// := db.AutoMigrate(&User{})
	if migrateErr != nil {
		return
	}
	r := gin.Default()
	// load status files
	//r.Static("/static", "templates/statics")
	////模板解析
	//r.LoadHTMLGlob("templates/**/*")
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")
	r.GET("/", index)
	r.GET("/download", downloadFile)
	r.GET("/delete", deleteFile)
	//r.POST("/upload", uploadFile)
	r.POST("/upload1", uploadFile1)
	_ = r.Run(":8081")
}

func index(c *gin.Context) {

	var uploadfile []UploadFile
	db.Model(&UploadFile{}).Find(&uploadfile)
	fmt.Println(uploadfile)
	c.HTML(http.StatusOK, "index.html", gin.H{
		"result": uploadfile,
	})
}

func downloadFile(c *gin.Context) {
	downloadId := c.Query("file_id")
	fmt.Println("download id :", downloadId)
	var downloadFile UploadFile
	db.Model(&UploadFile{}).Where("Id = ?", downloadId).Find(&downloadFile)
	fmt.Println(downloadFile.DownloadPath)
	_, errByOpenFile := os.Open(downloadFile.DownloadPath)
	if errByOpenFile != nil {
		//c.Redirect(http.StatusFound, "/404")
		c.JSON(200, gin.H{"1": "2"})

	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+downloadFile.FileName)
	c.Header("Content-Transfer-Encoding", "binary")
	c.File(downloadFile.DownloadPath)
	return
}

func deleteFile(c *gin.Context) {
	downloadId := c.Query("id")
	key := c.Query("key")
	if key != ACCESSKEY {
		//c.Redirect(http.StatusFound, "/404")
		c.JSON(200, gin.H{"status": "error", "data": "key error!"})
		return
	}
	var downloadFile UploadFile
	db.Model(&UploadFile{}).Where("Id = ?", downloadId).Find(&downloadFile)
	fmt.Println(downloadFile.DownloadPath)
	errByOpenFile := os.Remove(downloadFile.DownloadPath)
	if errByOpenFile != nil {
		//		//c.Redirect(http.StatusFound, "/404")
		c.JSON(200, gin.H{"status": "error"})
		return

	}
	db.Where("Id = ?", downloadId).Delete(&UploadFile{})
	c.JSON(200, gin.H{"status": "success", "data": "delete success"})
	return
}
func uploadFile1(c *gin.Context) {
	file, err := c.FormFile("file")
	key := c.PostForm("key")
	if err == nil && key == ACCESSKEY {

		fmt.Println("key:", key)
		// 判断文件是否为空
		if file == nil {
			c.JSON(http.StatusOK, gin.H{
				"status": "error",
				"data":   "file nil！",
			})
			return
		}
		// 判断文件大小
		if file.Size > 5000000000 || file.Size <= 10 {
			c.JSON(http.StatusOK, gin.H{
				"status": "error",
				"data":   "file size error！",
			})
			return
		}

		// 查询文件名是否已经存在
		var files1 []UploadFile
		db.Model(&UploadFile{}).Where("Hash = ?", MD5(file.Filename)).Find(&files1)
		if len(files1) > 0 {
			c.JSON(http.StatusOK, gin.H{
				"status": "error",
				"data":   "file exits！",
			})
			return
		}

		// 限制上传数量
		var downloadFile []UploadFile
		db.Model(&UploadFile{}).Find(&downloadFile)
		file_nums := len(downloadFile)
		//fmt.Println("file lens:", len(downloadFile))
		if file_nums > 50 {
			c.JSON(http.StatusOK, gin.H{
				"status": "error",
				"data":   "file nums error！",
			})
			return
		}
		dst := path.Join("./files", file.Filename)
		saveErr := c.SaveUploadedFile(file, dst)
		if saveErr == nil {

			uploadfile := UploadFile{
				FileName:       file.Filename,
				FileSize:       Decimal(float64(float64(file.Size) / (1024 * 1024))),
				Hash:           MD5(file.Filename),
				DownloadNumber: 0,
				DownloadPath:   dst,
			}
			//fmt.Println("size mb:", float64(file.Size/(1024*1024)))
			insertErr := db.Model(&UploadFile{}).Create(&uploadfile).Error
			if insertErr != nil {
				panic(insertErr)
			}
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status": "error",
			})
			return
		}

	} else {
		c.JSON(http.StatusOK, gin.H{
			"status": "key error！",
		})
		return
	}
}

// 浮点数保留3位小数
func Decimal(num float64) float64 {
	num, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", num), 64)
	return num
}
