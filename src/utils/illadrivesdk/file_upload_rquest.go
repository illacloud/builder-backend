package illadrivesdk

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/illacloud/builder-backend/src/utils/idconvertor"
)

// The request sample like:
// ```json
// {"resumable":true,"name":"lemmy.hjson","folderID":"ILAfx4p1C7cX","type":"file","size":590,"duplicationHandler":"manual","contentType":""}
// ```
type UploadFileRequest struct {
	Name               string `json:"name" validate:"required"`
	FolderID           string `json:"folderID" validate:"required"`
	Type               string `json:"type" validate:"oneof=file folder"`
	ContentType        string `json:"contentType"`
	Resumable          bool   `json:"resumable"`
	Size               int64  `json:"size"`
	DuplicationHandler string `json:"duplicationHandler" validate:"oneof=cover rename manual"`
	Cover              bool   `json:"-"`
}

func NewUploadFileRequest() *UploadFileRequest {
	return &UploadFileRequest{}
}

func NewUploadFileRequestByParam(resumable bool, name string, folderID string, fileType string, size int64, duplicationHandler string, contentType string) *UploadFileRequest {
	return &UploadFileRequest{
		Resumable:          resumable,
		Name:               name,
		FolderID:           folderID,
		Type:               fileType,
		Size:               size,
		DuplicationHandler: duplicationHandler,
		ContentType:        contentType,
	}
}

func (r *UploadFileRequest) ExportFileName() string {
	return r.Name
}

func (r *UploadFileRequest) ExportFileType() string {
	return r.Type
}

func (r *UploadFileRequest) ExportFileTypeInt() int {
	if r.Type == "file" {
		return 3
	}
	return 2
}

func (r *UploadFileRequest) ExportParentID() int {
	return idconvertor.ConvertStringToInt(r.FolderID)
}

func (r *UploadFileRequest) ExportFolderID() string {
	return r.FolderID
}

func (r *UploadFileRequest) ExportSize() int64 {
	return r.Size
}

func (r *UploadFileRequest) ExportMIMEType() string {
	return r.ContentType
}

func (r *UploadFileRequest) ExportDuplicationStrategy() string {
	return r.DuplicationHandler
}

func (r *UploadFileRequest) ExportResumable() bool {
	return r.Resumable
}

func (r *UploadFileRequest) SetNewFileName(newFileName string) {
	r.Name = newFileName
}

func (r *UploadFileRequest) SetNewParentID(newParentID int) {
	r.FolderID = idconvertor.ConvertIntToString(newParentID)
}

func (r *UploadFileRequest) GenerateNewFileName() string {
	base := filepath.Base(r.Name)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 3)
	for i := range b {
		b[i] = letterBytes[rd.Intn(len(letterBytes))]
	}

	return fmt.Sprintf("%s_%s%s", name, string(b), ext)
}

func (r *UploadFileRequest) SetDuplicationCover(cover bool) {
	r.Cover = cover
}

func (r *UploadFileRequest) ExportDuplicationCover() bool {
	return r.Cover
}

func (r *UploadFileRequest) ExportContentType() string {
	return r.ContentType
}
