package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadImage 上传图片
func UploadImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		const maxImages = 5
		const maxFileSize = 2 << 20 // 2MB
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无法解析表单数据", "success": false})
			return
		}
		files := form.File["images"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "未上传任何图片", "success": false})
			return
		}
		if len(files) > maxImages {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("最多上传 %d 张图片", maxImages), "success": false})
			return
		}
		var imageURLs []string
		uploadDir := "static/uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			log.Printf("创建上传目录失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		for _, file := range files {
			if file.Size > maxFileSize {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("图片 %s 超过2MB限制", file.Filename), "success": false})
				return
			}
			ext := strings.ToLower(filepath.Ext(file.Filename))
			if ext != ".jpg" && ext != ".png" {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("图片 %s 格式不支持，仅支持 jpg/png", file.Filename), "success": false})
				return
			}
			filename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), strings.TrimSuffix(file.Filename, ext), ext)
			dstPath := filepath.Join(uploadDir, filename)
			src, err := file.Open()
			if err != nil {
				log.Printf("打开上传文件失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
				return
			}
			defer src.Close()
			dst, err := os.Create(dstPath)
			if err != nil {
				log.Printf("创建目标文件失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
				return
			}
			defer dst.Close()
			if _, err := io.Copy(dst, src); err != nil {
				log.Printf("保存文件失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
				return
			}
			imageURLs = append(imageURLs, fmt.Sprintf("/static/uploads/%s", filename))
		}
		c.JSON(http.StatusOK, gin.H{"message": "图片上传成功", "image_urls": imageURLs})
	}
}

// DeleteImage 删除图片
func DeleteImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			ImageURL string `json:"image_url"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "success": false})
			return
		}
		if !strings.HasPrefix(request.ImageURL, "/static/uploads/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图片路径", "success": false})
			return
		}
		filename := strings.TrimPrefix(request.ImageURL, "/")
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "图片不存在", "success": false})
			return
		}
		if err := os.Remove(filename); err != nil {
			log.Printf("删除图片 %s 失败: %v", filename, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误", "success": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "图片删除成功"})
	}
}
