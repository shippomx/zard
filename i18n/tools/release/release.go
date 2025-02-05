package release

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/shippomx/zard/i18n/tools/release/client"
)

var (
	token     string
	targetDir string
	workSpace string
	repoSlug  string
)

func Init() *cobra.Command {
	releaseCmd := cobra.Command{
		Use:   "release",
		Short: "Release the i18n config to bitbucket downloads.",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Releasing I18N config to bitbucket...")
			fmt.Println("targetdir: " + targetDir)
			fmt.Println("workspace: " + workSpace)
			fmt.Println("repoSlug: " + repoSlug)
			return Release()
		},
	}
	releaseCmd.Flags().StringVarP(&token, "token", "t", "", "bitbucket token")
	releaseCmd.Flags().StringVarP(&targetDir, "dir", "d", "", "i18n config dir")
	releaseCmd.Flags().StringVarP(&workSpace, "workspace", "w", "gatebackend", "bitbucket workspace")
	releaseCmd.Flags().StringVarP(&repoSlug, "repo", "r", "i18n", "bitbucket repo slug")
	return &releaseCmd
}

func Release() error {
	filenname, err := ZipTarget(targetDir)
	fmt.Println(filenname)
	if err != nil {
		return err
	}
	cc := client.NewClient(workSpace, repoSlug, token)
	err = cc.CreateDownloads(filenname)
	if err != nil {
		return err
	}
	return nil
}

// 将 targetdir 文件夹压缩为 时间戳.zip
// 放到os.temp 文件夹 返回路径
func ZipTarget(targetdir string) (string, error) {
	zipPath := fmt.Sprintf("%s/%s.zip", os.TempDir(), time.Now().Format(time.RFC3339))
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(targetdir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(targetdir, filePath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		return "", err
	}

	return zipPath, nil
}
