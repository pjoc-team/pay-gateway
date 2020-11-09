package service

import (
	"encoding/json"
	"fmt"
	etcdfileutil "github.com/coreos/etcd/pkg/fileutil"
	"github.com/pjoc-team/tracing/logger"
	"io/ioutil"
	"os"
)

// Store storage for services
type Store interface {
	// Put put service
	Put(serviceName string, service *Service) error
	// Get get service name
	Get(serviceName string) (*Service, error)
}

// fileStore use file storage to implements the store interface
type fileStore struct {
	filePath     string
	lockFilePath string
	file         *os.File
	lockedFile   *etcdfileutil.LockedFile
}

// NewFileStore create file store
func NewFileStore(filePath string) (Store, error) {
	log := logger.Log()

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, etcdfileutil.PrivateFileMode)
	// file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err2 := file.Close()
		if err2 != nil {
			log.Error(err2.Error())
		}
	}()
	lockFilePath := fmt.Sprintf("%s.lock", filePath)

	fs := &fileStore{
		filePath:     filePath,
		lockFilePath: lockFilePath,
		file:         file,
	}

	return fs, nil
}

func (f *fileStore) lock() error {
	file, err := etcdfileutil.TryLockFile(
		f.lockFilePath, os.O_WRONLY|os.O_CREATE, 0777,
	)
	if err != nil {
		return err
	}
	f.lockedFile = file
	return nil
}
func (f *fileStore) unlock() error {
	if f.lockedFile == nil {
		return nil
	}
	err := os.Remove(f.lockFilePath)
	if err != nil {
		return err
	}
	return nil
}

func (f *fileStore) Put(serviceName string, service *Service) error {
	log := logger.Log()
	for {
		err := f.lock()
		if err != nil {
			log.Errorf(
				"failed to put service, "+
					"because lock failed serviceName: %v service: %#v", serviceName, service,
			)
			continue
		}
		break
	}
	defer func() {
		err := f.unlock()
		if err != nil {
			log.Errorf(
				"failed to put service, "+
					"because unlock failed serviceName: %v service: %#v", serviceName, service,
			)
		}
	}()
	file, err := os.OpenFile(f.filePath, os.O_RDWR|os.O_CREATE, etcdfileutil.PrivateFileMode)
	// file, err := os.Open(filePath)
	if err != nil {
		log.Errorf(
			"failed to put service, " +
				"because open file failed serviceName: %v service",
		)
		return err
	}
	defer func() {
		err2 := file.Close()
		if err2 != nil {
			log.Error(err2.Error())
		}
	}()

	services, err := f.readAll()
	if err != nil {
		return err
	}
	services[serviceName] = service
	marshal, err := json.Marshal(services)
	if err != nil {
		return err
	}
	_, err = file.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}

func (f *fileStore) Get(serviceName string) (*Service, error) {
	services, err2 := f.readAll()
	if err2 != nil {
		return nil, err2
	}
	return services[serviceName], nil
}

func (f *fileStore) readAll() (map[string]*Service, error) {
	raw, err := ioutil.ReadFile(f.filePath)
	if err != nil {
		return nil, err
	}
	services := make(map[string]*Service)
	if len(raw) == 0 {
		return services, nil
	}
	err = json.Unmarshal(raw, &services)
	if err != nil {
		return nil, err
	}
	return services, nil
}
