package repository

import (
	"os"
	"sync"
)

type FileRepository struct {
	filePath     string
	mxFileAccess sync.Mutex // защищаем доступ к файлу
}

func NewFileRepository(filePath string) *FileRepository {
	return &FileRepository{
		filePath: filePath,
	}
}

func (r *FileRepository) FilePath() string {
	if r == nil {
		return ""
	} else {
		return r.filePath
	}
}

func (r *FileRepository) SaveMetrics(data []byte) error {
	if r.filePath == "" {
		return nil
	}

	if len(data) == 0 {
		return ErrNoData
	}

	r.mxFileAccess.Lock()
	defer r.mxFileAccess.Unlock()

	if err := os.WriteFile(r.filePath, data, 0600); err != nil {
		return NewErrWriteFile(r.filePath, err)
	}

	return nil
}

func (r *FileRepository) LoadMetrics() ([]byte, error) {
	if r.filePath == "" {
		return make([]byte, 0), nil
	}

	r.mxFileAccess.Lock()
	defer r.mxFileAccess.Unlock()

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]byte, 0), nil
		}
		return nil, NewErrReadFile(r.filePath, err)
	}

	return data, nil

}
