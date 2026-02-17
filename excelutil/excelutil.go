package internal

import (
	"mime/multipart"

	"github.com/xuri/excelize/v2"
)

func OpenReader(fileHeader multipart.FileHeader) (*excelize.File, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}

	defer file.Close()

	//这里读取文件内容到f上之后就可以关闭file了
	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, err
	}
	return f, nil
}
