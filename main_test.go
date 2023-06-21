package main

import (
	"archive/zip"
	"bytes"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/disintegration/imaging"
	cleanerMock "github.com/rendau/kazan/internal/adapters/cleaner/mock"
	"github.com/rendau/kazan/internal/adapters/logger/zap"
	"github.com/rendau/kazan/internal/cns"
	"github.com/rendau/kazan/internal/domain/core"
	"github.com/rendau/kazan/internal/domain/errs"
	"github.com/rendau/kazan/internal/domain/types"
	"github.com/rendau/kazan/internal/domain/util"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const confPath = "test_conf.yml"
const testDirPath = "test_dir"
const imgMaxWidth = 1000
const imgMaxHeight = 1000

type fsItemSt struct {
	p  string
	c  string
	mt time.Time
}

var (
	app = struct {
		lg      *zap.St
		cleaner *cleanerMock.St
		core    *core.St
	}{}
)

func cleanTestDir() {
	err := filepath.Walk(testDirPath, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		if p == testDirPath {
			return nil
		}

		// app.lg.Infow("cleanTestDir walk", "path", p, "name", info.Name(), "mt", info.ModTime())

		if info.IsDir() {
			err = os.RemoveAll(p)
			if err != nil {
				return err
			}

			return filepath.SkipDir
		}

		return os.Remove(p)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	var err error

	err = os.MkdirAll(testDirPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	viper.SetConfigFile(confPath)
	_ = viper.ReadInConfig()

	viper.AutomaticEnv()

	app.lg, err = zap.New(
		"info",
		true,
		false,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer app.lg.Sync()

	app.cleaner = cleanerMock.New()

	app.core = core.New(
		app.lg,
		app.cleaner,
		testDirPath,
		imgMaxWidth,
		imgMaxHeight,
		"",
		0,
		[]string{},
		0,
		time.Minute,
		true,
	)

	// Start tests
	code := m.Run()

	// cleanTestDir()

	os.Exit(code)
}

func TestCreate(t *testing.T) {
	cleanTestDir()

	_, err := app.core.Static.Create("asd/"+cns.ZipDirNamePrefix+"_asd", "a.txt", bytes.NewBuffer([]byte("test_data")), false, false)
	require.NotNil(t, err)
	require.Equal(t, errs.BadDirName, err)

	_, err = app.core.Static.Create(cns.ZipDirNamePrefix+"_asd/asd", "a.txt", bytes.NewBuffer([]byte("test_data")), false, false)
	require.NotNil(t, err)
	require.Equal(t, errs.BadDirName, err)

	fPath, err := app.core.Static.Create("photos", "data.txt", bytes.NewBuffer([]byte("test_data")), false, false)
	require.Nil(t, err)

	fPathPrefix := "photos/" + time.Now().Format("2006/01/02") + "/"

	require.True(t, strings.HasPrefix(fPath, fPathPrefix))
	require.False(t, strings.Contains(strings.TrimPrefix(fPath, fPathPrefix), "/"))

	fName, _, fContent, err := app.core.Static.Get(fPath, &types.ImgParsSt{}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)
	require.Equal(t, "test_data", string(fContent))
	require.NotEmpty(t, fName)

	largeImg := imaging.New(imgMaxWidth+10, imgMaxHeight+10, color.RGBA{R: 0xaa, G: 0x00, B: 0x00, A: 0xff})
	require.NotNil(t, largeImg)

	largeImgBuffer := new(bytes.Buffer)

	err = imaging.Encode(largeImgBuffer, largeImg, imaging.JPEG)
	require.Nil(t, err)

	fPath, err = app.core.Static.Create("photos", "a.jpg", largeImgBuffer, true, false)
	require.Nil(t, err)

	_, _, fContent, err = app.core.Static.Get(fPath, &types.ImgParsSt{}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err := imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds := img.Bounds().Max
	require.Equal(t, imgMaxWidth+10, imgBounds.X)
	require.Equal(t, imgMaxHeight+10, imgBounds.X)

	largeImg = imaging.New(imgMaxWidth+10, imgMaxHeight+10, color.RGBA{R: 0xaa, G: 0x00, B: 0x00, A: 0xff})
	require.NotNil(t, largeImg)

	largeImgBuffer = new(bytes.Buffer)

	err = imaging.Encode(largeImgBuffer, largeImg, imaging.JPEG)
	require.Nil(t, err)

	fPath, err = app.core.Static.Create("photos", "a.jpg", largeImgBuffer, false, false)
	require.Nil(t, err)

	_, _, fContent, err = app.core.Static.Get(fPath, &types.ImgParsSt{}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err = imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds = img.Bounds().Max
	require.Equal(t, imgMaxWidth, imgBounds.X)
	require.Equal(t, imgMaxHeight, imgBounds.X)

	_, _, fContent, err = app.core.Static.Get(fPath, &types.ImgParsSt{Method: "fit", Width: imgMaxWidth - 10, Height: imgMaxHeight - 10}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err = imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds = img.Bounds().Max
	require.Equal(t, imgMaxWidth-10, imgBounds.X)
	require.Equal(t, imgMaxHeight-10, imgBounds.X)

	_, _, fContent, err = app.core.Static.Get(fPath, &types.ImgParsSt{Method: "fit", Width: imgMaxWidth + 10, Height: imgMaxHeight + 10}, false)
	require.Nil(t, err)
	require.NotNil(t, fContent)

	img, err = imaging.Decode(bytes.NewBuffer(fContent))
	require.Nil(t, err)

	imgBounds = img.Bounds().Max
	require.Equal(t, imgMaxWidth, imgBounds.X)
	require.Equal(t, imgMaxHeight, imgBounds.X)
}

func TestCreateZip(t *testing.T) {
	cleanTestDir()

	zipContentIsSame := func(a, b []fsItemSt) {
		require.Equal(t, len(a), len(b))

		for _, ai := range a {
			found := false

			for _, bi := range b {
				if bi.p == ai.p {
					found = true
					require.Equal(t, ai.c, bi.c)
					break
				}
			}

			require.True(t, found)
		}

		for _, bi := range b {
			found := false

			for _, ai := range a {
				if ai.p == bi.p {
					found = true
					require.Equal(t, bi.c, ai.c)
					break
				}
			}

			require.True(t, found)
		}
	}

	srcZipFiles := []fsItemSt{
		{p: "index.html", c: "some html content"},
		{p: "abc/file.txt", c: "file content"},
		{p: "abc/qwe/x.txt", c: "x content"},
		{p: "todo.txt", c: "todo content"},
	}

	zipBuffer, err := createZipArchive(srcZipFiles)
	require.Nil(t, err)

	_, err = app.core.Static.Create("zip/"+cns.ZipDirNamePrefix+"_asd", "a.zip", zipBuffer, false, true)
	require.NotNil(t, err)
	require.Equal(t, errs.BadDirName, err)

	_, err = app.core.Static.Create(cns.ZipDirNamePrefix+"_asd/zip", "a.zip", zipBuffer, false, true)
	require.NotNil(t, err)
	require.Equal(t, errs.BadDirName, err)

	fPath, err := app.core.Static.Create("zip", "a.zip", zipBuffer, false, true)
	require.Nil(t, err)
	require.True(t, strings.HasSuffix(fPath, "/"))

	for _, zp := range srcZipFiles {
		_, _, fContent, err := app.core.Static.Get(fPath+zp.p, &types.ImgParsSt{}, false)
		require.Nil(t, err)
		require.NotNil(t, fContent)
		require.Equal(t, zp.c, string(fContent))
	}

	fName, _, fContent, err := app.core.Static.Get(fPath, &types.ImgParsSt{}, false)
	require.Nil(t, err)
	require.Equal(t, "index.html", fName)
	require.NotNil(t, fContent)
	require.Equal(t, "some html content", string(fContent))

	fName, _, fContent, err = app.core.Static.Get(fPath, &types.ImgParsSt{}, true)
	require.Nil(t, err)
	require.True(t, strings.HasSuffix(fName, ".zip"))
	require.NotNil(t, fContent)

	resultZipFiles, err := extractZipArchive(fContent)
	require.Nil(t, err)
	zipContentIsSame(srcZipFiles, resultZipFiles)

	srcZipFiles = []fsItemSt{
		{p: "root/index.html", c: "some html content"},
		{p: "root/abc/file.txt", c: "file content"},
		{p: "root/abc/qwe/x.txt", c: "x content"},
		{p: "root/todo.txt", c: "todo content"},
	}

	zipBuffer, err = createZipArchive(srcZipFiles)
	require.Nil(t, err)

	fPath, err = app.core.Static.Create("zip", "a.zip", zipBuffer, false, true)
	require.Nil(t, err)
	require.True(t, strings.HasSuffix(fPath, "/"))

	for _, zp := range srcZipFiles {
		_, _, fContent, err := app.core.Static.Get(fPath+strings.TrimLeft(zp.p, "root/"), &types.ImgParsSt{}, false)
		require.Nil(t, err)
		require.NotNil(t, fContent)
		require.Equal(t, zp.c, string(fContent))
	}

	fName, _, fContent, err = app.core.Static.Get(fPath, &types.ImgParsSt{}, false)
	require.Nil(t, err)
	require.Equal(t, "index.html", fName)
	require.NotNil(t, fContent)
	require.Equal(t, "some html content", string(fContent))
}

// func TestClean(t *testing.T) {
// 	cleanTestDir()
//
// 	cleanTime := time.Now().AddDate(0, 0, -cns.CleanFileNotCheckPeriodDays-1)
//
// 	dirStructure := []fsItemSt{
// 		{p: "dir1", c: "", mt: cleanTime},
// 		{p: "dir2/file1.txt", c: "file1 content", mt: cleanTime},
// 		{p: "dir2/dir3/file2.txt", c: "file2 content", mt: cleanTime},
// 		{p: "file3.txt", c: "file3 content", mt: cleanTime},
// 		{p: "dir5/" + cns.ZipDirNamePrefix + "q/a/js.js", c: "content", mt: cleanTime},
// 		{p: "dir5/" + cns.ZipDirNamePrefix + "q/index.html", c: "content", mt: cleanTime},
// 		{p: "dir5/" + cns.ZipDirNamePrefix + "q/css.css", c: "content", mt: cleanTime},
// 		{p: "dir6/q" + cns.ZipDirNamePrefix + "/index.html", c: "content", mt: cleanTime},
// 	}
//
// 	err := makeDirStructure(testDirPath, dirStructure)
// 	require.Nil(t, err)
//
// 	compareDirStructure(t, testDirPath, dirStructure)
//
// 	checkedFiles := make([]string, 0)
//
// 	app.cleaner.SetHandler(func(pathList []string) []string {
// 		checkedFiles = append(checkedFiles, pathList...)
// 		return []string{}
// 	})
//
// 	app.core.Clean(0)
//
// 	dirStructure = dirStructure[1:]
//
// 	compareStringSlices(t, []string{
// 		"dir2/file1.txt",
// 		"dir2/dir3/file2.txt",
// 		"file3.txt",
// 		"dir5/" + cns.ZipDirNamePrefix + "q/",
// 		"dir6/q" + cns.ZipDirNamePrefix + "/index.html",
// 	}, checkedFiles)
//
// 	compareDirStructure(t, testDirPath, dirStructure)
//
// 	app.cleaner.SetHandler(func(pathList []string) []string {
// 		return []string{
// 			dirStructure[1].p,
// 		}
// 	})
//
// 	app.core.Clean(0)
//
// 	dirStructure = append(dirStructure[:1], dirStructure[2:]...)
//
// 	compareDirStructure(t, testDirPath, dirStructure)
//
// 	checkedFiles = make([]string, 0)
//
// 	app.cleaner.SetHandler(func(pathList []string) []string {
// 		checkedFiles = append(checkedFiles, pathList...)
// 		return pathList
// 	})
//
// 	err = os.Chtimes(filepath.Join(testDirPath, "file3.txt"), time.Now(), time.Now())
// 	require.Nil(t, err)
//
// 	err = os.Chtimes(filepath.Join(testDirPath, "dir5", cns.ZipDirNamePrefix+"q"), time.Now(), time.Now())
// 	require.Nil(t, err)
//
// 	app.core.Clean(0)
//
// 	dirStructure = []fsItemSt{
// 		{p: "file3.txt", c: "file3 content", mt: cleanTime},
// 		{p: "dir5/" + cns.ZipDirNamePrefix + "q/a/js.js", c: "content", mt: cleanTime},
// 		{p: "dir5/" + cns.ZipDirNamePrefix + "q/index.html", c: "content", mt: cleanTime},
// 		{p: "dir5/" + cns.ZipDirNamePrefix + "q/css.css", c: "content", mt: cleanTime},
// 	}
//
// 	app.lg.Info(checkedFiles)
//
// 	compareStringSlices(t, []string{
// 		"dir2/file1.txt",
// 		"dir6/q" + cns.ZipDirNamePrefix + "/index.html",
// 	}, checkedFiles)
//
// 	compareDirStructure(t, testDirPath, dirStructure)
//
// 	cleanTestDir()
//
// 	err = makeDirStructure(testDirPath, []fsItemSt{
// 		{p: "dir1", c: "", mt: cleanTime},
// 		{p: "dir2", c: "", mt: cleanTime},
// 		{p: "dir2/dir3", c: "", mt: cleanTime},
// 		{p: "dir2/dir3/dir4", c: "", mt: cleanTime},
// 		{p: "dir2/dir5/dir6", c: "", mt: cleanTime},
// 		{p: "dir2/dir5/file1.txt", c: "asd", mt: cleanTime},
// 	})
// 	require.Nil(t, err)
//
// 	app.cleaner.SetHandler(func(pathList []string) []string { return []string{} })
//
// 	app.core.Clean(0)
//
// 	compareDirStructure(t, testDirPath, []fsItemSt{
// 		{p: "dir2/dir5/file1.txt", c: "asd"},
// 	})
// }

func createZipArchive(items []fsItemSt) (*bytes.Buffer, error) {
	result := new(bytes.Buffer)

	zipWriter := zip.NewWriter(result)
	defer zipWriter.Close()

	for _, item := range items {
		f, err := zipWriter.Create(item.p)
		if err != nil {
			log.Fatal(err)
		}
		_, err = f.Write([]byte(item.c))
		if err != nil {
			log.Fatal(err)
		}
	}

	return result, nil
}

func extractZipArchive(data []byte) ([]fsItemSt, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	result := make([]fsItemSt, 0)

	fileHandler := func(f *zip.File) error {
		if f.FileInfo().IsDir() {
			return nil
		}

		srcFile, err := f.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		srcFileDataRaw, err := ioutil.ReadAll(srcFile)
		if err != nil {
			return err
		}

		result = append(result, fsItemSt{p: f.Name, c: string(srcFileDataRaw), mt: f.Modified})

		return nil
	}

	for _, f := range reader.File {
		err = fileHandler(f)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func compareDirStructure(t *testing.T, dirPath string, items []fsItemSt) {
	diskItems := make([]fsItemSt, 0)

	err := filepath.Walk(dirPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil {
			return nil
		}

		if p == dirPath {
			return nil
		}

		relP, err := filepath.Rel(dirPath, p)
		if err != nil {
			return err
		}

		relUrlP := util.ToUrlPath(relP)

		if info.IsDir() {
			diskItems = append(diskItems, fsItemSt{p: relUrlP})
		} else {
			fileDataRaw, err := ioutil.ReadFile(p)
			if err != nil {
				return err
			}

			diskItems = append(diskItems, fsItemSt{p: relUrlP, c: string(fileDataRaw)})
		}

		return nil
	})
	require.Nil(t, err)

	for _, dItem := range diskItems {
		found := false
		dItemIsDir := path.Ext(dItem.p) == ""

		for _, item := range items {
			if dItemIsDir {
				if strings.Contains(item.p, dItem.p) {
					found = true
					break
				}
			} else {
				if item.p == dItem.p {
					require.Equal(t, item.c, dItem.c)
					found = true
					break
				}
			}
		}

		require.True(t, found, "Item not found %s", dItem.p)
	}

	for _, item := range items {
		found := false

		for _, dItem := range diskItems {
			if item.p == dItem.p {
				require.Equal(t, item.c, dItem.c)
				found = true
				break
			}
		}

		require.True(t, found, "Item not found %s", item.p)
	}
}

func makeDirStructure(parentDirPath string, items []fsItemSt) error {
	var err error

	for _, item := range items {
		fsAbsPath := filepath.Join(util.ToFsPath(parentDirPath), util.ToFsPath(item.p))

		if path.Ext(item.p) == "" { // dir
			err = os.MkdirAll(fsAbsPath, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(fsAbsPath), os.ModePerm)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(fsAbsPath, []byte(item.c), os.ModePerm)
			if err != nil {
				return err
			}

			if !item.mt.IsZero() {
				err = os.Chtimes(fsAbsPath, item.mt, item.mt)
				if err != nil {
					return err
				}

				if ind := strings.Index(fsAbsPath, "/"+cns.ZipDirNamePrefix); ind > -1 {
					err = os.Chtimes(fsAbsPath[:ind+len("/"+cns.ZipDirNamePrefix)+1], item.mt, item.mt)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func compareStringSlices(t *testing.T, a, b []string) {
	require.Equal(t, len(a), len(b))

	for _, aI := range a {
		found := false

		for _, bI := range b {
			if bI == aI {
				found = true
				break
			}
		}

		require.True(t, found, "String not found %q", aI)
	}

	for _, bI := range b {
		found := false

		for _, aI := range a {
			if aI == bI {
				found = true
				break
			}
		}

		require.True(t, found, "String not found %q", bI)
	}
}
