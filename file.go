package rollingfile

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	currentTime    = time.Now
	defaultMaxAge  = 3600        //1 hour
	defaultMaxSize = 1024_000_00 //100 Mb
	defaultSuffix  = ""
)

type options struct {
	maxSize int
	maxAge  int
	suffix  string
}

type Option func(*options)

func MaxSize(maxSize int) Option {
	return func(opts *options) { opts.maxSize = maxSize }
}

func MaxAge(maxAge int) Option {
	return func(opts *options) { opts.maxAge = maxAge }
}

func Suffix(suffix string) Option {
	return func(opts *options) { opts.suffix = suffix }
}

type RollingFile struct {
	filename string
	maxSize  int
	maxAge   int
	suffix   string

	current *os.File
	size    int64
	ctime   time.Time
	mu      sync.Mutex
}

// New returns *RollingFile implements io.WriteCloer
// *RollingFile rotate file with both size limit or writing time limit.
// The current active filename is always the name you specified.
// Option:
//   MaxSize(bytes int) file size threshold value for rolling, default: 102400000(100Mb)
//   MaxAge(seconds int) file writing time threshold value for rolling, default: 3600(1 hour)
//   Suffix(str string) rotated filename format: {base_name}-{timestr}{suffix}.{extension}, default empty
func New(filename string, opts ...Option) (*RollingFile, error) {
	mopts := &options{
		maxSize: defaultMaxSize,
		maxAge:  defaultMaxAge,
		suffix:  defaultSuffix,
	}
	for _, opt := range opts {
		opt(mopts)
	}

	return &RollingFile{
		filename: filename,
		maxSize:  mopts.maxSize,
		maxAge:   mopts.maxAge,
		suffix:   mopts.suffix,
	}, nil
}

func (f *RollingFile) Write(p []byte) (n int, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	writeSize := len(p)

	if f.current == nil {
		if err = f.openCurrent(writeSize); err != nil {
			return 0, err
		}
	}

	if currentTime().Sub(f.ctime) > time.Second*time.Duration(f.maxAge) {
		if err := f.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = f.current.Write(p)
	f.size += int64(n)

	if f.size >= int64(f.maxSize) {
		f.rotate()
	}

	return n, err
}

// openCurrent set a valid file handler to current
func (f *RollingFile) openCurrent(writeSize int) error {
	filename := f.filename

	info, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return f.openNew()
	}

	if err != nil {
		return fmt.Errorf("get file info err:%s", err)
	}

	if info.Size()+int64(writeSize) >= int64(f.maxSize) {
		return f.rotate()
	}

	fh, err := os.OpenFile(f.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, info.Mode())
	if err != nil {
		return fmt.Errorf("open file err:%s", err)
	}

	f.current = fh
	f.size = info.Size()
	f.ctime = info.ModTime()

	return nil
}

// openNew create current file when not exists
func (f *RollingFile) openNew() error {
	dir := filepath.Dir(f.filename)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("can't make directories for rollingfile:%s", err)
	}

	mode := os.FileMode(0644)
	fh, err := os.OpenFile(f.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, mode)

	if err != nil {
		return fmt.Errorf("can't make rollingfile:%s", err)
	}

	f.current = fh
	f.size = 0
	f.ctime = currentTime()

	return nil
}

// rotate rename current file to rotatedName
// then reopen current file that is empty file
func (f *RollingFile) rotate() error {
	if err := f.close(); err != nil {
		return err
	}
	rotatedName := f.rotatedName()
	if err := os.Rename(f.filename, rotatedName); err != nil {
		return fmt.Errorf("can't rename file:%s", err)
	}

	return f.openNew()
}

// rotatedName get rotated filename
// format: base_file_dir/base_file_name-{timestamp}{suffix}.ext
func (f *RollingFile) rotatedName() string {
	dir := filepath.Dir(f.filename)
	filename := filepath.Base(f.filename)
	ext := filepath.Ext(filename)
	prefix := filename[:len(filename)-len(ext)]
	timestamp := currentTime().Format("20060102150405")
	i := 0
	for {
		dupIndex := ""
		if i > 0 {
			dupIndex = fmt.Sprintf(".%d", i)
		}
		filename := filepath.Join(
			dir,
			fmt.Sprintf("%s-%s%s%s%s", prefix, timestamp, f.suffix, dupIndex, ext),
		)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return filename
		}
		i++
	}
}

func (f *RollingFile) close() error {
	if f.current == nil {
		return nil
	}
	err := f.current.Close()
	f.current = nil
	return err
}

func (f *RollingFile) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.close()
}
