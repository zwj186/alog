package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/zwj186/alog/log"
)

type _FileConfig struct {
	Size       int64
	Path       string
	RetainDay  int
	GCInterval time.Duration
	ChildTmpl  *template.Template
	NameTmpl   *template.Template
	TimeTmpl   *template.Template
	MsgTmpl    *template.Template
}

// NewFileStore 创建新的FileStore实例
func NewFileStore(config log.FileConfig) log.LogStore {
	var (
		size      = config.FileSize
		fpath     = config.FilePath
		childpath = config.ChildPathTmpl
		filename  = config.FileNameTmpl
		timeTmpl  = config.Item.TimeTmpl
		msgTmpl   = config.Item.Tmpl
		interval  = config.GCInterval
	)
	if size == 0 {
		size = log.DefaultFileSize
	}
	if fpath == "" {
		fpath = log.DefaultFilePath
	}
	if !filepath.IsAbs(fpath) {
		fpath, _ = filepath.Abs(fpath)
	}
	if filename == "" {
		filename = log.DefaultFileNameTmpl
	}
	if timeTmpl == "" {
		timeTmpl = log.DefaultTimeTmpl
	}
	if msgTmpl == "" {
		msgTmpl = log.DefaultMsgTmpl
	}
	if interval == 0 {
		interval = log.DefaultFileGCInterval
	}
	if l := len(fpath); l > 0 && fpath[l-1] == '/' {
		fpath = fpath[:l-1]
	}

	cfg := &_FileConfig{
		Size:       size * 1024,
		Path:       fpath,
		ChildTmpl:  template.Must(template.New("").Parse(childpath)),
		NameTmpl:   template.Must(template.New("").Parse(filename)),
		TimeTmpl:   template.Must(template.New("").Parse(timeTmpl)),
		MsgTmpl:    template.Must(template.New("").Parse(msgTmpl)),
		RetainDay:  config.RetainDay,
		GCInterval: time.Duration(interval) * time.Minute,
	}
	fs := &FileStore{config: cfg, fileMap: make(map[string]*FileUnit)}

	// 创建日志目录
	if err := fs.createFolder(fs.config.Path); err != nil {
		panic("创建目录发生错误：" + err.Error())
	}

	if config.RetainDay > 0 {
		// 清理过期的文件
		go func() {
			fs.gc()
		}()
	}

	return fs
}

// FileUnit 提供文件管理
type FileUnit struct {
	fileName string
	file     *os.File
	size     int64
}

// FileStore 提供文件日志存储
type FileStore struct {
	config  *_FileConfig
	fileMap map[string]*FileUnit
}

// 执行文件清理
func (fs *FileStore) gc() {
	ct := time.Now()
	err := filepath.Walk(fs.config.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.ModTime().Before(ct.AddDate(0, 0, -fs.config.RetainDay)) {
			os.Remove(path)
		}
		return nil
	})

	if err != nil {
		fmt.Println("FileStore GC Error:", err.Error())
	}

	time.AfterFunc(fs.config.GCInterval, fs.gc)
}

func (fs *FileStore) createFolder(folder string) error {
	//folder := fs.config.Path
	_, err := os.Stat(folder)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(folder, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs *FileStore) changeName(name, ext string) string {
	var number int
	prefix := fmt.Sprintf("%s/%s", fs.config.Path, name)
	err := filepath.Walk(fs.config.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasPrefix(path, prefix) {
			return nil
		}
		number++
		return nil
	})
	if err != nil {
		return fmt.Sprintf("%s%s", name, ext)
	}
	return fmt.Sprintf("%s_%d%s", name, number, ext)
}

func (fs *FileStore) rename(fileName string) (err error) {
	ext := filepath.Ext(fileName)
	prefix := fileName[:len(fileName)-len(ext)]
	err = os.Rename(fmt.Sprintf("%s/%s", fs.config.Path, fileName), fmt.Sprintf("%s/%s", fs.config.Path, fs.changeName(prefix, ext)))
	return
}

func (fs *FileStore) getFile(item *log.LogItem) (fileUnit *FileUnit, err error) {
	childPath := log.ParseName(fs.config.ChildTmpl, item)
	fileName := log.ParseName(fs.config.NameTmpl, item)
	if fileName == "" {
		fileName = fmt.Sprintf("unknown.%s.log", item.Time.Format("20060102"))
	}
	if len(childPath) > 0 {
		fileName = childPath + "/" + fileName
	}
	fileUnit, ok := fs.fileMap[fileName]
	if ok {
		if fileUnit.file != nil {
			finfo, err := fileUnit.file.Stat()
			if err != nil || (finfo.Size() > 0 && finfo.Size() >= fs.config.Size) {
				fileUnit.file.Close()
				fileUnit.file = nil
				fs.rename(fileUnit.fileName)
			} else if err == nil {
				fileUnit.file.Close()
				fileUnit.file = nil
			} else {
				return fileUnit, err
			}
		} else if fileUnit.size > 0 && fileUnit.size >= fs.config.Size {
			fs.rename(fileUnit.fileName)
		}
	} else {
		fileUnit = &FileUnit{fileName: fileName}
		fs.fileMap[fileUnit.fileName] = fileUnit
	}

	if fileUnit.file == nil {
		if len(childPath) > 0 {
			fs.createFolder(fs.config.Path + "/" + childPath)
		}
		file, err := os.OpenFile(fmt.Sprintf("%s/%s", fs.config.Path, fileUnit.fileName), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
		if err != nil {
			return fileUnit, err
		}
		fileUnit.file = file
		return fileUnit, err
	}
	return fileUnit, err
}

func (fs *FileStore) Store(item *log.LogItem) (err error) {
	fileUnit, err := fs.getFile(item)
	if err != nil {
		return
	}

	logInfo := log.ParseLogItem(fs.config.MsgTmpl, fs.config.TimeTmpl, item)
	_, err = fileUnit.file.WriteString(logInfo)
	if err != nil {
		return
	}

	finfo, err := fileUnit.file.Stat()
	if err == nil {
		fileUnit.size = finfo.Size()
	}
	if err != nil || finfo.Size() >= fs.config.Size {
		fileUnit.file.Close()
		fileUnit.file = nil
	}
	return
}

func (fs *FileStore) Close() (err error) {
	for fileName, fileUnit := range fs.fileMap {
		if fileUnit.file != nil {
			fileUnit.file.Close()
			fileUnit.file = nil
			if fileUnit.size >= fs.config.Size {
				fs.rename(fileUnit.fileName)
			}
		}
		delete(fs.fileMap, fileName)
	}
	return
}
