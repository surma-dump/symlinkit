package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	help   *bool   = flag.Bool("h", false, "Show this help")
	force  *bool   = flag.Bool("f", false, "Overwrite existing symlinks")
	clean  *bool   = flag.Bool("c", false, "Clean dead links")
	prefix *string = flag.String("prefix", "", "Prefix to prepend to all symlinks")
)

func main() {
	log.SetPrefix("symlinkit: ")
	defer func() {
		if x := recover(); x != nil {
			e, ok := x.(error)
			if ok {
				log.Fatalf("%s\n", e.Error())
			} else {
				log.Fatalf("Unknown error\n")
			}
		}
	}()

	flag.Parse()
	if *help {
		flag.PrintDefaults()
		log.Fatalf("Usage: symlinkit [-f] [-c] [-h] <src directory> <dst directory>\n")
	}
	srcdir, dstdir := flag.Arg(0), flag.Arg(1)
	if !(isDir(srcdir) && isDir(dstdir)) {
		log.Fatalf("Both parameters have to be directories")
	}
	srcdir = appendSeparator(srcdir)
	dstdir = appendSeparator(dstdir)
	makeSymlinks(srcdir, dstdir)
	if *clean {
		cleanSymlinks(dstdir)
	}
}

func makeSymlinks(srcdir, dstdir string) {
	l := linker{srcdir, dstdir}
	filepath.Walk(srcdir, func(path string, f os.FileInfo, e error) error {
		// Symlinks should be resolved
		f, e = os.Stat(path)
		if e != nil {
			return filepath.SkipDir
		}
		if f.IsDir() {
			if !l.VisitDir(path, f) {
				return filepath.SkipDir
			}
		} else {
			l.VisitFile(path, f)
		}
		return nil
	})
}

type linker struct {
	srcdir, dstdir string
}

func (m linker) VisitDir(path string, f os.FileInfo) bool {
	if path == m.srcdir {
		return true
	}
	suffix := path[len(m.srcdir):]
	if !isDir(m.dstdir + suffix) {
		e := os.Mkdir(m.dstdir+suffix, 0755)
		check(e)
	}
	return true
}

func (m linker) VisitFile(path string, f os.FileInfo) {
	suffix := path[len(m.srcdir):]
	dirpath, file := filepath.Split(suffix)
	srcfile := m.srcdir + suffix
	dstfile := m.dstdir + dirpath + *prefix + file
	if *force {
		os.Remove(dstfile)
	}
	e := os.Symlink(srcfile, dstfile)
	check(e)
}

func cleanSymlinks(path string) {
	c := cleaner(path)
	filepath.Walk(path, func(path string, f os.FileInfo, e error) error {
		if f.IsDir() {
			if !c.VisitDir(path, f) {
				return filepath.SkipDir
			}
		} else {
			c.VisitFile(path, f)
		}
		return nil
	})
}

type cleaner string

func (cleaner) VisitDir(path string, f os.FileInfo) bool {
	return true
}

func (cleaner) VisitFile(path string, f os.FileInfo) {
	if f.Mode() & os.ModeSymlink != 0 {
		target, e := os.Readlink(path)
		if e != nil || (!isFile(target) && !isDir(target)) {
			os.Remove(path)
		}
	}
}

func appendSeparator(path string) string {
	sep := string(filepath.Separator)
	if !strings.HasSuffix(path, sep) {
		return path + sep
	}
	return path
}

func isDir(path string) bool {
	fi, e := os.Stat(path)
	return e == nil && fi.IsDir()
}

func isFile(path string) bool {
	fi, e := os.Stat(path)
	if e != nil {
		return false
	}
	isPipe := (fi.Mode() & os.ModeNamedPipe) != 0
	isSocket := (fi.Mode() & os.ModeSocket) != 0
	isSymlink := (fi.Mode() & os.ModeSymlink) != 0
	return !isPipe && !isSocket && !isSymlink
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
