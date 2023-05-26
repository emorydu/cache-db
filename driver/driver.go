package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jcelliott/lumber"
	"os"
	"path/filepath"
	"sync"
)

const (
	Version = "1.0.0"
	dotJson = ".json"
	dotTmp  = ".tmp"
)

type driver struct {
	sync.Mutex
	mutexes map[string]*sync.Mutex
	dir     string
	log     Logger
}

var ErrDirIsEmpty = errors.New("dir is not empty")

var (
	ErrCollectionUnableWrite = errors.New("missing collection - no place to save record")
	ErrResourceUnableWrite   = errors.New("missing collection - unable to save record (no name)")

	ErrCollectionUnableRead = errors.New("missing collection - unable to read")
	ErrResourceUnableRead   = errors.New("missing resource - unable to read record (no name)")
)

func New(dir string, options *Options) (*driver, error) {

	_, _ = fmt.Fprintf(os.Stdout, `
-----------------------------------------
| System Name: Simple Key-Value Cache-DB
| Version:     %s				
-----------------------------------------
`, Version)
	if dir == "" {
		return nil, ErrDirIsEmpty
	}

	dir = filepath.Clean(dir)
	opts := &Options{}
	if options != nil {
		opts = options
	}
	if options == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	driver := &driver{
		mutexes: make(map[string]*sync.Mutex, 1000),
		dir:     dir,
		log:     opts.Logger,
	}

	if _, err := os.Stat(dir); err != nil {
		opts.Logger.Debug("Using '%s' (database already exists), Make sure your data is backed up\n", dir)
		return driver, nil
	}
	opts.Debug("Creating the database at '%s'...\n", dir)
	return driver, os.MkdirAll(dir, 0755)
}

func (d *driver) Write(collection, resource string, v any) error {
	if len(collection) == 0 {
		return ErrCollectionUnableWrite
	}
	if len(resource) == 0 {
		return ErrResourceUnableWrite
	}

	// Gets a collection lock
	// Here the collection is equivalent to a disk file, avoiding concurrent access.
	mu := d.getOrCreateMutex(collection)
	mu.Lock()
	defer mu.Unlock()

	dir := filepath.Join(d.dir, collection)
	finalPath := filepath.Join(dir, resource+dotJson)
	tmpPath := finalPath + dotTmp

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	b = append(b, byte('\n'))

	if err = os.WriteFile(tmpPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, finalPath)
}

func (d *driver) Read(collection, resource string, v any) error {
	if len(collection) == 0 {
		return ErrCollectionUnableRead
	}
	if len(resource) == 0 {
		return ErrResourceUnableRead
	}

	record := filepath.Join(d.dir, collection, resource)
	if _, err := stat(record); err != nil {
		return err
	}

	if b, err := os.ReadFile(record + dotJson); err != nil {
		return err
	} else {
		return json.Unmarshal(b, &v)
	}
}

func (d *driver) ReadAll(collection string) ([]string, error) {
	if len(collection) == 0 {
		return nil, ErrCollectionUnableRead
	}

	dir := filepath.Join(d.dir, collection)
	if _, err := stat(dir); err != nil {
		return nil, err
	}

	files, _ := os.ReadDir(dir)

	var records []string
	for _, v := range files {
		b, err := os.ReadFile(filepath.Join(dir, v.Name()))
		if err != nil {
			return nil, err
		}

		records = append(records, string(b))
	}
	return records, nil
}

func (d *driver) Delete(collection, resource string) error {
	path := filepath.Join(collection, resource)
	mu := d.getOrCreateMutex(collection)
	mu.Lock()
	defer mu.Unlock()

	dir := filepath.Join(d.dir, path)

	switch f, err := stat(dir); {
	case f == nil, err != nil:
		return fmt.Errorf("unable to find file or directory named %v\n", path)
	case f.Mode().IsDir():
		return os.RemoveAll(dir)
	case f.Mode().IsRegular():
		return os.RemoveAll(dir + dotJson)
	}

	return nil
}

type Options struct {
	Logger
}

func (d *driver) getOrCreateMutex(collection string) *sync.Mutex {
	d.Lock()
	defer d.Unlock()

	m, ok := d.mutexes[collection]
	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}

	return m
}

func stat(path string) (info os.FileInfo, err error) {
	if info, err = os.Stat(path); os.IsNotExist(err) {
		info, err = os.Stat(path + dotJson)
	}
	return
}
