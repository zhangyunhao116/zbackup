package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhangyunhao116/wyhash"
)

func main() {
	if len(os.Args) < 2 {
		logrus.Fatal("Need target directory!")
	}
	curdir, err := os.Getwd()
	if err != nil {
		logrus.Fatal("Get work directory:", err.Error())
	}
	targetdir := path.Join(curdir, os.Args[1])
	targetfile, err := os.Open(targetdir)
	if err != nil {
		logrus.Fatal("Get target file:", err.Error())
	}
	targetfilestat, err := targetfile.Stat()
	if err != nil {
		logrus.Fatal("Get target file stat:", err.Error())
	}
	if !targetfilestat.IsDir() {
		logrus.Fatal(targetfile.Name(), " is not a directory")
	}
	CompressDir(targetdir, curdir)
}

func CompressFileName(targetdir string) string {
	_, dirname := path.Split(targetdir)
	now := time.Now()
	y, month, d := now.Date()
	h, m, s := now.Clock()
	timeprefix := fmt.Sprintf("%d-%d-%d-%d-%d-%d", y, int(month), d, h, m, s)
	filename := fmt.Sprintf("%s-%s", timeprefix, dirname)
	return filename
}

func CompressDir(targetdir, workdir string) {
	pre := "CompressDir"
	filename := CompressFileName(targetdir)
	// Compress.
	execCommandPrintOnlyFailed(pre,
		fmt.Sprintln("tar --use-compress-program=zstd -cf", `"`+filename+`"`, `"`+targetdir+`"`),
	)
	// Sum.
	tarfile, err := os.Open(path.Join(workdir, filename))
	if err != nil {
		logrus.Fatal(fmt.Sprintf("Open target file failed(%s):", path.Join(workdir, filename)), err.Error())
	}
	digest := wyhash.NewDefault()
	buffer := make([]byte, 1024)
	for {
		n, err := tarfile.Read(buffer)
		if err == io.EOF {
			digest.Write(buffer[:n])
			break
		} else if err != nil {
			panic(err)
		}
		digest.Write(buffer[:n])
	}
	hashsum := fmt.Sprintf("%016x", digest.Sum64())
	// Rename.
	finalFileName := filename + "-wyf1-" + hashsum + ".zbackup"
	execCommandPrintOnlyFailed("rename", "mv "+filename+" "+finalFileName)
}

func execCommand(prefix, cmd string) (string, error) {
	var stderr bytes.Buffer
	command := exec.Command("bash", "-c", cmd)
	command.Stderr = &stderr
	out, err := command.Output()
	if err != nil {
		return stderr.String(), errors.New(prefix + ": " + err.Error())
	}
	return string(out), nil
}

func execCommandPrint(prefix, cmd string) (string, error) {
	out, err := execCommand(prefix, cmd)
	if err != nil {
		logrus.Warningf(`exec "%s" error: `, prefix)
	}
	print(string(out))
	return out, err
}

func execCommandPrintOnlyFailed(prefix, cmd string) (string, error) {
	out, err := execCommand(prefix, cmd)
	if err != nil {
		logrus.Warningf(`exec "%s" error: `, prefix)
		print(string(out))
	}
	return out, err
}
