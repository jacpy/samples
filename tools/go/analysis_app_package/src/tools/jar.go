package tools

import (
	"io/ioutil"
	"strings"
	"log"
	"io"
	"archive/zip"
)

type PkgInfo struct {
	MaxLevel int
	root node
	Pkg []string
}

type node struct {
	next []*node
	value string
	count int
}

var pkgData map[string]*PkgInfo = make(map[string]*PkgInfo)

func AnalysisJar(p string) error {
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return nil
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".jar") {
			analysisJarPackageName(f.Name())
		} else if strings.HasSuffix(f.Name(), ".aar") {

		}
	}

	return nil
}

func analysisJarPackageName(p string)  {
	log.Println(p)
}

func ParsePackageName(apkName, className string)  {
	if len(apkName) == 0 || len(className) == 0 {
		return
	}

	pkg, exists := pkgData[apkName]
	if !exists {
		log.Println(apkName, "is not exists.")
		//pkg = &PkgInfo{&nodeList{make([]*node, 4), ""}, make([]string, 10)}
		pkg = &PkgInfo{0, node{make([]*node, 0, 4), "", 0}, make([]string, 0, 10)}
		pkgData[apkName] = pkg
	}

	names := strings.Split(className, "/")
	length := len(names)
	if pkg.MaxLevel < length {
		pkg.MaxLevel = length
	}

	var n *node = nil
	parent := &pkg.root
	list := pkg.root.next
	log.Println("--->length:", len(list))
	for _, value := range names {
		for _, nl := range list {
			//log.Println("nl:", nl.value, ", value:", value, ", compare:", nl.value == value, ", count: ", nl.count)
			if nl.value == value {
				n = nl
				list = nl.next
				parent = n
				nl.count++
				break
			}
		}

		//log.Println("***>length:", len(list), n, ", idx:", idx)
		if n == nil {
			// 没有找到值相同的节点
			nodeItem := &node{make([]*node, 0, 1), value, 1}
			n = nodeItem
			parent.next = append(parent.next, nodeItem)
			//log.Println("###list:")
			//for _, l := range list {
			//	log.Print(l.value, ",", l.count)
			//}
			//
			//log.Println("###")
			parent = nodeItem
			list = nodeItem.next
		}

		n = nil
	}
}

func AnalysisPkg(writer io.Writer, level int, ignoreClass []string) map[string]*PkgInfo {
	log.Println("===>map len:", len(pkgData))
	for apk, value := range pkgData {
		log.Println(apk)
		writer.Write([]byte("--"))
		writer.Write([]byte(apk))
		writer.Write([]byte("\n"))
		t := level
		if value.MaxLevel < level {
			t = value.MaxLevel - 1
		}

		//log.Println("@@@len:", len(value.root.next), value.root.next, t)
		if t <= 0 {
			continue
		}

		analysisPkg(writer, value.root.next, "", 1, t, ignoreClass)
	}
	//for _, value := range pkgData {
	//	count += len(value.allClass)
	//	for _, clsName := range value.allClass {
	//		if contains(clsName, value.Pkg) {
	//			continue
	//		}
	//
	//		pkgs := strings.Split(clsName, "/")
	//		log.Println(clsName)
	//	}
	//}

	return pkgData
}

func analysisPkg(writer io.Writer, n []*node, header string, idx, level int, ignoreClass []string) {
	//log.Println("===>length:", len(n.next))
	if idx <= level {
		for _, item := range n {
			if len(item.next) == 0 {
				// 没有下一级了
				printClassName(writer, header, ignoreClass)
				break
			}

			if len(header) == 0 {
				analysisPkg(writer, item.next, item.value, idx + 1, level, ignoreClass)
			} else {
				analysisPkg(writer, item.next, header + "." + item.value, idx + 1, level, ignoreClass)
			}
		}
	} else {
		printClassName(writer, header, ignoreClass)
	}
}

func printClassName(writer io.Writer, header string, ignoreClass []string) {
	isSystem := false
	for _, s:= range ignoreClass {
		if strings.Replace(s, "/", ".", -1) == header {
			//log.Println("~~~~~s: ", s, " isEquals: ", header)
			isSystem = true
			break
		}
	}

	if !isSystem {
		log.Println(header)
		writer.Write([]byte(header))
		writer.Write([]byte("\n"))
	}
}

// 从jar中解析出包名
func ParsePackageNameFromJar(level int, jars... string) []string {
	pkg := make([]string, 0)
	for _, jar := range jars {
		rc, err := zip.OpenReader(jar)
		if err != nil {
			log.Fatal(err)
		}

		packages := make([]string, 0)
		for _, file := range rc.File {
			name := file.Name
			if file.FileInfo().IsDir() || !strings.HasSuffix(name, ".class") {
				continue
			}

			// 有class文件，则记为包名
			if strings.Count(name, "/") <= level {
				name = strings.Replace(name[:strings.LastIndex(name, "/")], "/", ".", -1)
			} else {
				idx := 1
				idx = strings.IndexFunc(name, func(r rune) bool {
					if r == '/' {
						if idx == level {
							return true
						}

						idx++
					}

					return false
				})

				name = strings.Replace(name[:idx], "/", ".", -1)
			}

			existsPkg := false
			for _, p := range packages {
				if strings.HasPrefix(p, name) {
					existsPkg = true
					break
				}
			}

			if !existsPkg {
				packages = append(packages, name)
			}

		}

		rc.Close()
		pkg = append(pkg, packages...)
	}
	return pkg
}

func hasSubClass(name, sub string) bool {
	return strings.Count(name, "/") == strings.Count(sub, "/") && strings.HasSuffix(sub, ".class")
}
