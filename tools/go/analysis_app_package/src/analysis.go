package main

import (
	"log"
	"os"
	"archive/zip"
	"strings"
	"io"
	"path/filepath"
	"io/ioutil"
	"os/exec"
	"bytes"
	"bufio"
	"ctp/eimLibAnalysis/src/tools"
)

// 解析dex文件中的类名
func parseDex(classWriter io.Writer, apkPath, dexPath string) {
	p, err := os.Getwd()
	//cmd := exec.Command(filepath.Join(p, "dexdump.exe"), dexPath, "|", "grep", "Class descriptor")
	//dump := fmt.Sprintf(`dexdump.exe %s | grep "Class descriptor"`, dexPath)
	//log.Println(dump)
	dumpAll := exec.Command(filepath.Join(p, "dexdump.exe"), dexPath)
	classDesc := exec.Command("grep", "-E", "Class descriptor|type")
	classDesc.Stdin, err = dumpAll.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}

	b := new(bytes.Buffer)
	classDesc.Stdout = b
	//cmd := exec.Command(filepath.Join(p, "dexdump.exe"), `dexdump.exe D:/temp/log/0607/classes.dex | grep "Class descriptor"`)
	//b, err := cmd.CombinedOutput()
	err = classDesc.Start()
	if err != nil {
		log.Println(err)
		return
	}

	dumpAll.Run()
	classDesc.Wait()
	//b, err := classDesc.CombinedOutput()
	r := bufio.NewReader(b)
	for line, _, _ := r.ReadLine(); len(line) != 0; line, _, _ = r.ReadLine() {
		str := string(line)
		log.Println(str)
		names := findClassName(str)
		if len(names) == 0 {
			continue
		}

		for _, name := range names {
			log.Println(name)
			classWriter.Write([]byte(name))
			classWriter.Write([]byte("\n"))
			tools.ParsePackageName(filepath.Base(apkPath), name)
		}
	}

	//log.Println(string(b.Bytes()))
}

// 从字符串中解析出类名在"L"和";"之间的
func findClassName(str string) []string {
	name := make([]string, 0)
	for {
		idx := strings.Index(str, "L")
		if idx == -1 {
			break
		}

		idx2 := strings.Index(str, ";")
		if idx2 == -1 {
			break
		}

		s := str[idx + 1: idx2]
		name = append(name, s)
		str = str[idx2 + 1:]
	}

	return name
}

// 解析APK将APK中包含的DEX或者APK文件释放出来
func parseAPK(classWriter io.Writer, p string) {
	r, err := zip.OpenReader(p)
	if err != nil {
		log.Fatal(err)
	}

	defer r.Close()
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".apk") || strings.HasSuffix(f.Name, ".dex") {
			log.Println(f.Name)
			rc, err := f.Open()
			if err != nil {
				log.Println(err)
				continue
			}

			log.Println("--->base:", filepath.Base(p))
			fp := filepath.Join(filepath.Dir(p), "tmp", filepath.Base(p) + "." + f.Name)
			err = releaseFile(fp, rc)
			rc.Close()
			if err != nil {
				log.Println(err)
				continue
			}

			if strings.HasSuffix(fp, ".apk") {
				parseAPK(classWriter, fp)
			} else if strings.HasSuffix(fp, ".dex") {
				classWriter.Write([]byte("--"))
				classWriter.Write([]byte(fp))
				classWriter.Write([]byte("\n"))
				parseDex(classWriter, p, fp)
			}
		}
	}
}

// 解压APK中的文件
func releaseFile(p string, rc io.ReadCloser) error {
	log.Println(p)
	os.MkdirAll(filepath.Dir(p), os.ModePerm)
	b, err := ioutil.ReadAll(rc)
	if err != nil {
		log.Println("read failed: ", p)
		return err
	}

	err = ioutil.WriteFile(p, b, os.ModePerm)
	if err != nil {
		log.Println("write failed: ", p)
		return err
	}

	return nil
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	args := os.Args
	dir, _ := os.Getwd()
	file, err := os.OpenFile(filepath.Join(dir, "pkgs.txt"), os.O_RDWR | os.O_CREATE | os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer file.Close()
	cls, err := os.OpenFile(filepath.Join(dir, "classes.txt"), os.O_RDWR | os.O_CREATE | os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}

	defer cls.Close()
	if len(args) >= 2 {
		log.Println("start parsing ", args[1])
		parseAPK(cls, args[1])
		jars := args[2:]
		log.Println("--->jars: ", jars)
		names := tools.ParsePackageNameFromJar(3, jars...)
		tools.AnalysisPkg(file, 3, names)
		//for _, name := range names {
		//	log.Println("###>", name)
		//}
	} else {
		log.Println("invalid args length: ", len(args) - 1)
	}
}
