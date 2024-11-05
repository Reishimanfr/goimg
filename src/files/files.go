package files

import (
	"bash06/goimg/src/database"
	"bash06/goimg/src/flags"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"gorm.io/gorm"
)

var (
	ErrUserNotVerified    = errors.New("User isn't verified")
	ErrUserNotExist       = errors.New("User doesn't exist")
	ErrInvalidManagerMode = errors.New("Invalid file manager mode provided")
	ErrMkdirPermission    = errors.New("Insufficient permissions to create directory")
	ErrMkdirInvalid       = errors.New("Invalid mkdir arguments")
	ErrFileOpen           = errors.New("Failed to open file")
	ErrFileCreate         = errors.New("Failed to create file")
	ErrFileCopy           = errors.New("Failed to copy file contents")
	ManagerModeAWS        = "aws"     // TODO
	ManagerModeOnDisk     = "on-disk" // TODO
	ManagerModeWebDav     = "web-dav" // TODO
	ManagerModeRemote     = "remote"  // TODO
)

type FileManager struct {
	Mode string
	Db   *gorm.DB
}

type FileInfo struct {
	OwnerId   string
	Header    multipart.FileHeader
	MimeType  string
	ExpiresAt int64
}

func New(mode string, db *gorm.DB) *FileManager {
	return &FileManager{
		Mode: mode,
		Db:   db,
	}
}

// Saves the provided file depending on the selected file manager mode.
// If mode is invalid it'll return an error and abort the process
func (f *FileManager) Save(file *FileInfo) error {
	var user *database.UserRecord
	var err error

	// We don't have to check if guest uploads are enabled here since we're already checking for that
	// in the uploads route func
	if file.OwnerId != "" {
		if err := f.Db.Model(&database.UserRecord{}).Where("user_id = ?", file.OwnerId).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return ErrUserNotExist
			}

			return fmt.Errorf("failed to query database: %v", err)
		}

		if !user.IsVerified {
			return ErrUserNotVerified
		}
	}

	switch f.Mode {
	case "aws":
		err = f.processS3Upload(file, user)

	case "on-disk":
		err = f.processLocalUpload(file, user)

	// case "web-dav":

	// case "remote":
	// 	{
	// 	}

	default:
		return ErrInvalidManagerMode

	}

	return err
}

func (f *FileManager) processS3Upload(file *FileInfo, user *database.UserRecord) error {
	return nil
}

func (f *FileManager) processLocalUpload(file *FileInfo, user *database.UserRecord) error {
	var userDirPath string

	if user == nil {
		userDirPath = filepath.Join(flags.BasePath, "guest_uploads")
	} else {
		userDirPath = filepath.Join(flags.BasePath, user.UUID)
	}

	err := os.MkdirAll(userDirPath, os.ModeDir)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return ErrMkdirPermission
		}

		if errors.Is(err, fs.ErrInvalid) {
			return ErrMkdirInvalid
		}

		if !errors.Is(err, fs.ErrExist) {
			return err
		}
	}

	var existingFiles []string
	err = f.Db.Model(&database.UserRecord{}).Where("owner_id = ?").Select("filename").Find(&existingFiles).Error
	if err != nil {
		return err
	}

	i := 0

	for {
		if !slices.Contains(existingFiles, file.Header.Filename) {
			break
		}
		i++

		idxString := strconv.Itoa(i)
		file.Header.Filename = file.Header.Filename + "_(" + idxString + ")"

	}

	srcFile, err := file.Header.Open()
	if err != nil {
		return ErrFileOpen
	}

	defer srcFile.Close()

	filePath := filepath.Join(userDirPath, file.Header.Filename)

	dest, err := os.Create(filePath)
	if err != nil {
		return ErrFileCreate
	}

	defer dest.Close()

	writer := bufio.NewWriter(dest)
	defer writer.Flush()

	_, err = io.Copy(writer, srcFile)
	if err != nil {
		return ErrFileCopy
	}

	record := &database.FileRecord{
		CreatedAt: time.Now().Unix(),
		OwnerId:   file.OwnerId,
		ExpiresAt: file.ExpiresAt,
		MimeType:  file.MimeType,
		Location:  filePath,
		Filename:  file.Header.Filename,
	}

	return f.Db.Model(&database.UserRecord{}).Create(record).Error
}
