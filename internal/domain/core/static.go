package core

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rendau/dop/dopErrs"

	"github.com/rendau/kazan/internal/cns"
	"github.com/rendau/kazan/internal/domain/errs"
	"github.com/rendau/kazan/internal/domain/types"
)

type Static struct {
	r *St
}

func NewStatic(r *St) *Static {
	return &Static{
		r: r,
	}
}

func (c *Static) Create(reqDir string, reqFileName string, reqFile io.Reader, noCut bool, unZip bool) (string, error) {
	reqDirUrlPath := util.ToUrlPath(reqDir)

	if strings.Contains("/"+reqDirUrlPath, "/"+cns.ZipDirNamePrefix) {
		return "", errs.BadDirName
	}

	if strings.HasPrefix("/"+reqDirUrlPath, "/"+cns.KvsDirNamePrefix) {
		return "", errs.BadDirName
	}

	dateUrlPath := util.GetDateUrlPath()

	absFsDirPath := filepath.Join(c.r.dirPath, util.ToFsPath(reqDir), util.ToFsPath(dateUrlPath))

	err := os.MkdirAll(absFsDirPath, os.ModePerm)
	if err != nil {
		c.r.lg.Errorw("Fail to create dirs", err)
		return "", err
	}

	reqFileExt := strings.ToLower(filepath.Ext(reqFileName))

	var targetFsPath string
	var isZipDir bool

	if unZip && reqFileExt == ".zip" {
		targetFsPath, err = os.MkdirTemp(absFsDirPath, cns.ZipDirNamePrefix+"*")
		if err != nil {
			c.r.lg.Errorw("Fail to create temp-dir", err)
			return "", err
		}

		err = c.r.Zip.Extract(reqFile, targetFsPath)
		if err != nil {
			return "", err
		}

		isZipDir = true
	} else {
		targetFsPath, err = func() (string, error) {
			f, err := os.CreateTemp(absFsDirPath, "*"+reqFileExt)
			if err != nil {
				c.r.lg.Errorw("Fail to create temp-file", err)
				return "", err
			}
			defer f.Close()

			_, err = io.Copy(f, reqFile)
			if err != nil {
				c.r.lg.Errorw("Fail to copy data", err)
				return "", err
			}

			return f.Name(), nil
		}()
		if err != nil {
			return "", err
		}

		if !noCut {
			err = c.r.Img.Handle(targetFsPath, nil, &types.ImgParsSt{
				Method: "fit",
				Width:  c.r.imgMaxWidth,
				Height: c.r.imgMaxHeight,
			})
			if err != nil {
				return "", err
			}
		}
	}

	fileFsRelPath, err := filepath.Rel(c.r.dirPath, targetFsPath)
	if err != nil {
		c.r.lg.Errorw("Fail to get relative path", err, "path", targetFsPath, "base", c.r.dirPath)
		return "", err
	}

	fileUrlRelPath := util.ToUrlPath(fileFsRelPath)

	if isZipDir {
		fileUrlRelPath += "/"
	}

	return fileUrlRelPath, nil
}

func (c *Static) Get(reqPath string, imgPars *types.ImgParsSt, download bool) (string, time.Time, []byte, error) {
	var err error

	cKey := c.r.Cache.GenerateKey(reqPath, imgPars, download)

	if name, modTime, content := c.r.Cache.GetAndRefresh(cKey); content != nil {
		return name, modTime, content, nil
	}

	reqFsPath := util.ToFsPath(reqPath)
	absFsPath := filepath.Join(c.r.dirPath, reqFsPath)

	name := ""
	modTime := time.Now()
	content := make([]byte, 0)

	fInfo, err := os.Stat(absFsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			c.r.lg.Errorw("Fail to get stat of file", err, "f_path", absFsPath)
		}
		return "", modTime, nil, dopErrs.ObjectNotFound
	}

	if !download {
		modTime = fInfo.ModTime()
	}

	if fInfo.IsDir() {
		dirName := filepath.Base(absFsPath)

		if strings.HasPrefix(dirName, cns.ZipDirNamePrefix) {
			if download {
				archiveBuffer, err := c.r.Zip.CompressDir(absFsPath)
				if err != nil {
					return "", modTime, nil, err
				}

				return "archive.zip", modTime, archiveBuffer.Bytes(), nil
			} else if strings.HasSuffix(reqPath, "/") {
				absFsPath = filepath.Join(absFsPath, "index.html")
				name = "index.html"
				imgPars.Reset()
			} else {
				return "", modTime, nil, dopErrs.ObjectNotFound
			}
		} else {
			return "", modTime, nil, dopErrs.ObjectNotFound
		}
	} else {
		_, name = filepath.Split(absFsPath)
	}

	for _, p := range c.r.wMarkDirPaths {
		if strings.HasPrefix(reqFsPath, p) {
			imgPars.WMark = true
			break
		}
	}

	if !imgPars.IsEmpty() {
		buffer := new(bytes.Buffer)

		err = c.r.Img.Handle(absFsPath, buffer, imgPars)
		if err != nil {
			return "", modTime, nil, err
		}

		if buffer.Len() > 0 {
			content = buffer.Bytes()
		}
	}

	if len(content) == 0 {
		content, err = ioutil.ReadFile(absFsPath)
		if err != nil {
			c.r.lg.Errorw("Fail to read file", err, "f_path", absFsPath)
			return "", modTime, nil, err
		}
	}

	c.r.Cache.Set(cKey, name, modTime, content)

	return name, modTime, content, nil
}
