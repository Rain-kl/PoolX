package controller

import (
	"github.com/gin-gonic/gin"
	"poolx/internal/model"
	"poolx/internal/service"
	"strconv"
)

func GetAllFiles(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	files, err := service.ListFiles(p)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, files)
}

func SearchFiles(c *gin.Context) {
	keyword := c.Query("keyword")
	files, err := service.SearchFiles(keyword)
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	respondSuccess(c, files)
}

func UploadFile(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		respondFailure(c, err.Error())
		return
	}
	err = service.SaveUploadedFiles(form.File["file"], service.FileUploadInput{
		Description: c.PostForm("description"),
		Uploader:    c.GetString("username"),
		UploaderID:  c.GetInt("id"),
	}, c.SaveUploadedFile)
	if err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelError, "file upload failed | username="+c.GetString("username")+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}
	_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelInfo, "file upload succeeded | username="+c.GetString("username"))
	respondSuccessMessage(c, "")
}

func DeleteFile(c *gin.Context) {
	fileIdStr := c.Param("id")
	fileId, err := strconv.Atoi(fileIdStr)
	if err != nil || fileId == 0 {
		respondBadRequest(c, "无效的参数")
		return
	}
	if err = service.DeleteFileByID(fileId); err != nil {
		_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelError, "file delete failed | username="+c.GetString("username")+" | file_id="+fileIdStr+" | reason="+err.Error())
		respondFailure(c, err.Error())
		return
	}
	_ = service.AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelInfo, "file delete succeeded | username="+c.GetString("username")+" | file_id="+fileIdStr)
	respondSuccessMessage(c, "文件删除成功")
}

func DownloadFile(c *gin.Context) {
	path := c.Param("file")
	fullPath, err := service.ResolveDownloadPath(path)
	if err != nil {
		c.Status(403)
		return
	}
	c.File(fullPath)
	go func() {
		service.RecordFileDownload(path)
	}()
}
