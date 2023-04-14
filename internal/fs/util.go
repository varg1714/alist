package fs

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	stdpath "path"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func ClearCache(path string) {
	storage, actualPath, err := op.GetStorageAndActualPath(path)
	if err != nil {
		return
	}
	op.ClearCache(storage, actualPath)
}

func containsByName(files []model.Obj, file model.Obj) bool {
	for _, f := range files {
		if f.GetName() == file.GetName() {
			return true
		}
	}
	return false
}

var httpClient = &http.Client{}

func getFileStreamFromLink(file model.Obj, link *model.Link) (*model.FileStream, error) {
	var rc io.ReadCloser
	mimetype := utils.GetMimeType(file.GetName())
	if link.Data != nil {
		rc = link.Data
	} else if link.FilePath != nil {
		// create a new temp symbolic link, because it will be deleted after upload
		newFilePath := stdpath.Join(conf.Conf.TempDir, fmt.Sprintf("%s-%s", uuid.NewString(), file.GetName()))
		err := utils.SymlinkOrCopyFile(*link.FilePath, newFilePath)
		if err != nil {
			return nil, err
		}
		f, err := os.Open(newFilePath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open file %s", *link.FilePath)
		}
		rc = f
	} else {
		req, err := http.NewRequest(http.MethodGet, link.URL, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create request for %s", link.URL)
		}
		for h, val := range link.Header {
			req.Header[h] = val
		}
		res, err := httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get response for %s", link.URL)
		}
		mt := res.Header.Get("Content-Type")
		if mt != "" && strings.ToLower(mt) != "application/octet-stream" {
			mimetype = mt
		}
		rc = res.Body

		m3u8Stream := []string{"application/vnd.apple.mpegurl", "application/x-mpegURL"}
		if utils.SliceContains(m3u8Stream, mt) {

			tempFile, err := os.CreateTemp("", "output.*.mp4")
			if err != nil {
				log.Error("Failed to create temp file:", err)
			}

			// 构建FFmpeg的命令行参数字符串
			cmdArgs := []string{"-i", link.URL, "-f", "mp4", tempFile.Name(), "-nostdin", "-y"}

			// 创建FFmpeg命令
			cmd := exec.Command("ffmpeg", cmdArgs...)

			// 获取FFmpeg的标准错误输出管道
			stderr, err := cmd.StderrPipe()
			if err != nil {
				log.Fatal("Failed to get FFmpeg stderr pipe:", err)
			}

			err = cmd.Start()
			if err != nil {
				log.Error("Failed to start FFmpeg:", err)
			}

			// 将FFmpeg的错误输出内容通过标准输出打印
			go func() {
				_, err := io.Copy(os.Stdout, stderr)
				if err != nil {
					log.Error("Failed to copy FFmpeg stderr to stdout:", err.Error())
				}
			}()

			err = cmd.Wait()
			if err != nil {
				log.Error("FFmpeg command execution failed:", err)
			}

			// 将FFmpeg的输出内容写入到io.ReadCloser的网络流中
			rc, err = os.Open(tempFile.Name())
			if err != nil {
				log.Error("open temp file error:", err)
			}

			log.Info("FFmpeg command execution completed successfully!")
		}

	}
	// if can't get mimetype, use default application/octet-stream
	if mimetype == "" {
		mimetype = "application/octet-stream"
	}
	stream := &model.FileStream{
		Obj:        file,
		ReadCloser: rc,
		Mimetype:   mimetype,
	}
	return stream, nil
}
