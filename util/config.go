package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
)

func LoadConfig(fileName string, v interface{}) error {
	if _, err := os.Stat(fileName); err != nil {
		return err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	if err = json.NewDecoder(file).Decode(v); err != nil {
		return err
	}
	return nil
}

func SaveConfig(fileName string, v interface{}) error {
	tmpFileName := fileName + ".tmp"

	f, err := os.Create(tmpFileName)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(f).Encode(v); err != nil {
		f.Close()
		return err
	}
	f.Close()

	if _, err = os.Stat(fileName); err == nil {
		if err = os.Remove(fileName); err != nil {
			return err
		}
	}

	if err := os.Rename(tmpFileName, fileName); err != nil {
		return err
	}

	return nil
}

func ConfigExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}

func RemoveConfig(fileName string) error {
	if _, err := Execute("rm", []string{"-f", fileName}); err != nil {
		return err
	}
	return nil
}

func ListConfigIDs(root, prefix, suffix string) ([]string, error) {
	pattern := path.Join(root, fmt.Sprintf("%s*%s", prefix, suffix))
	out, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return []string{}, nil
	}
	for i := range out {
		out[i] = path.Base(out[i])
	}
	return ExtractNames(out, prefix, suffix)
}

type ObjectOperations interface {
	ConfigFile() (string, error)
}

func getObjectOps(obj interface{}) (ObjectOperations, error) {
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("BUG: Non-pointer was passed in")
	}
	t := reflect.TypeOf(obj).Elem()
	ops, ok := obj.(ObjectOperations)
	if !ok {
		return nil, fmt.Errorf("BUG: %v doesn't implement necessary methods for accessing object", t)
	}
	return ops, nil
}

func ObjectConfig(obj interface{}) (string, error) {
	ops, err := getObjectOps(obj)
	if err != nil {
		return "", err
	}
	config, err := ops.ConfigFile()
	if err != nil {
		return "", err
	}
	return config, nil
}

func ObjectLoad(obj interface{}) error {
	config, err := ObjectConfig(obj)
	if err != nil {
		return err
	}
	if !ConfigExists(config) {
		return fmt.Errorf("Cannot find object config %v", config)
	}
	if err := LoadConfig(config, obj); err != nil {
		return err
	}
	return nil
}

func ObjectExists(obj interface{}) (bool, error) {
	config, err := ObjectConfig(obj)
	if err != nil {
		return false, err
	}
	return ConfigExists(config), nil
}

func ObjectSave(obj interface{}) error {
	config, err := ObjectConfig(obj)
	if err != nil {
		return err
	}
	return SaveConfig(config, obj)
}

func ObjectDelete(obj interface{}) error {
	config, err := ObjectConfig(obj)
	if err != nil {
		return err
	}
	return RemoveConfig(config)
}
