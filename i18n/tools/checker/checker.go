package checker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/spf13/cobra"
)

var (
	jsonFile []string
	dirFile  []string
	Inited   bool
)

func Init() *cobra.Command {
	checkCmd := cobra.Command{
		Use:   "check",
		Short: "Check if the i18n config is valid.",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			Inited = true
		},
	}
	checkCmd.Flags().StringArrayVarP(&jsonFile, "file", "f", []string{}, "i18n config file")
	checkCmd.Flags().StringArrayVarP(&dirFile, "dir", "d", []string{}, "i18n config dir")
	checkCmd.MarkFlagsOneRequired("file", "dir")
	return &checkCmd
}

func CheckFile() error {
	for _, file := range jsonFile {
		fmt.Println("Checking file: " + file)
		fileInfo, err := os.Stat(file)
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			return errors.New(file + " is dir")
		}
		err = CheckJosnFileValid(file)
		if err != nil {
			return err
		}

	}
	return nil
}

func CheckDir() error {
	for _, dir := range dirFile {
		dirInfo, err := os.Stat(dir)
		if err != nil {
			return err
		}
		if !dirInfo.IsDir() {
			return errors.New(dir + " is not dir")
		}
		fileList := make([]string, 0)
		// 遍历目录下的所有文件 和子目录下的文件
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			logger.Infof("file: %s", path)
			fileList = append(fileList, path)
			return nil
		})
		if err != nil {
			return err
		}

		for _, filePath := range fileList {
			logger.Info("Checking file: " + filePath)
			err = CheckJosnFileValid(filePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Check() error {
	if err := CheckFile(); err != nil {
		return err
	}
	if err := CheckDir(); err != nil {
		return err
	}
	logger.Info("check success")
	return nil
}

// check josnFile format is valid
func CheckJosnFileValid(file string) error {
	logger.Info("Checking JSON file: " + file)
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	var m map[string]string
	err = json.Unmarshal(content, &m)
	if err != nil {
		return err
	}
	logger.Infof("JSON file %s is valid", file)

	err = CheckJSONForDuplicateKeys(bytes.NewReader(content))
	if err != nil {
		return errors.Join(fmt.Errorf("error while checking JSON: %s", file), err)
	}

	err = CheckJSONForDuplicateValues(bytes.NewReader(content))
	if err != nil {
		logger.Infof("warning while checking JSON: %s", err.Error())
	}
	return nil
}

func CheckJSONForDuplicateKeys(reader io.Reader) error {
	logger.Info("Checking JSON for duplicate keys...")
	decoder := json.NewDecoder(reader)
	keyMap := make(map[string]bool)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error while decoding JSON: %v", err)
		}

		switch t := token.(type) {
		case json.Delim:
			if t == '{' {
				// 开始新的对象，清空 keyMap
				keyMap = make(map[string]bool)
			}
		case string:
			if keyMap[t] {
				return fmt.Errorf("duplicate key found: %s", t)
			}
			keyMap[t] = true

			// 跳过键对应的值
			if !decoder.More() {
				return fmt.Errorf("missing value for key '%s'", t)
			}
			_, err := decoder.Token()
			if err != nil {
				return fmt.Errorf("error while decoding JSON value: %v", err)
			}
		}
	}

	return nil
}

func CheckJSONForDuplicateValues(reader io.Reader) error {
	logger.Info("Checking JSON for duplicate values...")
	decoder := json.NewDecoder(reader)
	valueMap := make(map[string]bool)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error while decoding JSON: %v", err)
		}

		switch t := token.(type) {
		case json.Delim:
			if t == '{' {
				// 开始新的对象，清空 valueMap
				valueMap = make(map[string]bool)
			}
		case string:
			// 获取键对应的值
			if !decoder.More() {
				return errors.New("missing value for key '" + t + "'")
			}
			valueToken, err := decoder.Token()
			if err != nil {
				return errors.New("error while decoding JSON value: " + err.Error())
			}
			value, ok := valueToken.(string)
			if !ok {
				return errors.New("value is not a string")
			}

			if valueMap[value] {
				return errors.New("duplicate value found: " + value)
			}
			valueMap[value] = true
		}
	}

	return nil
}
