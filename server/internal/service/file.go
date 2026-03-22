package service

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"poolx/internal/model"
	"poolx/internal/pkg/common"
	"poolx/internal/pkg/utils"
	"strings"
	"time"
)

type FileUploadInput struct {
	Description string
	Uploader    string
	UploaderID  int
}

func ListFiles(page int) ([]*model.File, error) {
	if page < 0 {
		page = 0
	}
	return model.GetAllFiles(page*common.ItemsPerPage, common.ItemsPerPage)
}

func SearchFiles(keyword string) ([]*model.File, error) {
	return model.SearchFiles(keyword)
}

func SaveUploadedFiles(files []*multipart.FileHeader, input FileUploadInput, saveUploadedFile func(*multipart.FileHeader, string) error) error {
	if len(files) == 0 {
		return fmt.Errorf("请先选择要上传的文件")
	}

	description := input.Description
	if description == "" {
		description = "无描述信息"
	}
	uploader := input.Uploader
	if uploader == "" {
		uploader = "访客用户"
	}
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	for _, file := range files {
		filename := filepath.Base(file.Filename)
		ext := filepath.Ext(filename)
		link := utils.GetUUID() + ext
		savePath := filepath.Join(common.UploadPath, link)
		if err := saveUploadedFile(file, savePath); err != nil {
			return err
		}

		fileObj := &model.File{
			Description: description,
			Uploader:    uploader,
			UploadTime:  currentTime,
			UploaderId:  input.UploaderID,
			Link:        link,
			Filename:    filename,
		}
		if err := fileObj.Insert(); err != nil {
			return err
		}
	}

	return nil
}

func DeleteFileByID(fileID int) error {
	fileObj := &model.File{Id: fileID}
	model.DB.Where("id = ?", fileID).First(&fileObj)
	if fileObj.Link == "" {
		return fmt.Errorf("文件不存在！")
	}
	return fileObj.Delete()
}

func ResolveDownloadPath(path string) (string, error) {
	fullPath := filepath.Join(common.UploadPath, path)
	if !strings.HasPrefix(fullPath, common.UploadPath) {
		return "", fmt.Errorf("invalid file path")
	}
	return fullPath, nil
}

func RecordFileDownload(link string) {
	model.UpdateDownloadCounter(link)
}
